package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
)

const (
	DBTimeout   = 120 // database query timeout
	qMetricsTbl = `
		CREATE TABLE IF NOT EXISTS metrics 	(
			id      INTEGER GENERATED ALWAYS AS IDENTITY,
			updated_ts TIMESTAMPTZ NOT NULL DEFAULT current_timestamp(3),
			name    VARCHAR NOT NULL,
			type    VARCHAR NOT NULL,
			f_value DOUBLE PRECISION,
			i_value INTEGER
		);	
	`
	qMetricsIdx = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_metrics_name_type
			ON metrics (name, type);
	`
	qFuncSetUpdatedTS = `
		CREATE OR REPLACE FUNCTION trg_set_updated_ts() RETURNS TRIGGER AS $$
		BEGIN
		   IF row(NEW.*) IS DISTINCT FROM row(OLD.*) THEN
			  NEW.updated_ts = current_timestamp(3); 
			  RETURN NEW;
		   ELSE
			  RETURN OLD;
		   END IF;
		END;
		$$ LANGUAGE PLPGSQL;
	`
	qMetricsTrg = `
		DROP TRIGGER IF EXISTS trg_metrics_updated ON metrics;
		CREATE TRIGGER trg_metrics_updated
		BEFORE UPDATE 
		ON metrics
		FOR EACH ROW
		EXECUTE FUNCTION trg_set_updated_ts();
	`
)

type DBStorage struct {
	db *sql.DB
}

// New is a DBStorage constructor
func New(ctx context.Context, uri string) (*DBStorage, error) {
	dbs := new(DBStorage)
	var err error
	dbs.db, err = sql.Open("pgx", uri)
	if err != nil {
		// Error here instead of Fatal to let server work without db to pass tests 10[ab]
		logger.Log.Error().Err(err).Msg("unable to setup database connection")
		return nil, err
	}
	if err = dbs.initDB(ctx); err != nil {
		// Error here instead of Fatal to let server work without db to pass tests 10[ab]
		logger.Log.Error().Err(err).Msg("unable to init database")
		return nil, err
	}
	return dbs, nil
}

func (dbs *DBStorage) RestoreStorage(ctx context.Context, s store.Storage) {
	logger.Log.Debug().Msg("starting storage restore")
	metrics := dbs.getMetrics(ctx)
	for _, m := range metrics {
		switch m.Type {
		case models.Gauge:
			s.SetGauge(m.Name, *m.FValue)
		case models.Counter:
			s.DelCounter(m.Name)
			s.IncCounter(m.Name, *m.IValue)
		}
	}
	logger.Log.Debug().Msg("done storage restore")
}
func (dbs *DBStorage) SaveMetric(ctx context.Context, m models.Metrics) {
	dbs.updateMetric(ctx, m)
}
func (dbs *DBStorage) SaveStorage(ctx context.Context, s store.Storage) {
	logger.Log.Debug().Msg("saving store to database")
	snap := s.Snapshot()
	for _, m := range snap {
		dbs.updateMetric(ctx, m)
	}
}

func (dbs *DBStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout*time.Second)
	defer cancel()
	return dbs.db.PingContext(ctx)
}

// getMetrics selects all metrics from database
func (dbs *DBStorage) getMetrics(ctx context.Context) []models.Metrics {
	var res []models.Metrics
	rows, err := dbs.db.QueryContext(ctx, `SELECT name, type, f_value, i_value FROM metrics`)
	if err != nil {
		logger.Log.Error().Err(err).Msg("unable to get all metrics from db")
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var (
			m      models.Metrics
			iValue sql.NullInt64
			fValue sql.NullFloat64
		)
		if err = rows.Scan(&m.Name, &m.Type, &fValue, &iValue); err != nil {
			logger.Log.Warn().Err(err).Msg("unable parse metric from db")
			break
		}
		if fValue.Valid {
			m.FValue = &fValue.Float64
		}
		if iValue.Valid {
			m.IValue = &iValue.Int64
		}
		res = append(res, m)
	}
	if err = rows.Err(); err != nil {
		logger.Log.Warn().Err(err).Msg("error while getting metrics from db")
	}
	return res
}

// updateMetric inserts metric in database, updates metric value if metric already exists
func (dbs *DBStorage) updateMetric(ctx context.Context, m models.Metrics) {
	var err error
	switch m.Type {
	case models.Gauge:
		_, err = dbs.db.ExecContext(ctx, `
			INSERT INTO metrics (name,type,f_value) VALUES ($1,$2,$3)
			ON CONFLICT (name,type)
			DO UPDATE SET f_value = excluded.f_value;
    	`,
			m.Name,
			m.Type,
			*m.FValue,
		)
	case models.Counter:
		_, err = dbs.db.ExecContext(ctx, `
			INSERT INTO metrics (name,type,i_value) VALUES ($1,$2,$3)
			ON CONFLICT (name,type)
			DO UPDATE SET i_value = excluded.i_value;
    	`,
			m.Name,
			m.Type,
			*m.IValue,
		)
	}
	if err != nil {
		logger.Log.Error().Err(err).Msgf("unable to update metric '%+v'", m)
		return
	}
	logger.Log.Debug().Msgf("metric '%+v' saved to db", m)
}

// initDB creates necessary database entities: tables, indexes, etc...
func (dbs *DBStorage) initDB(ctx context.Context) error {
	if _, err := dbs.db.ExecContext(ctx, qMetricsTbl); err != nil {
		return err
	}
	if _, err := dbs.db.ExecContext(ctx, qMetricsIdx); err != nil {
		return err
	}
	if _, err := dbs.db.ExecContext(ctx, qFuncSetUpdatedTS); err != nil {
		return err
	}
	if _, err := dbs.db.ExecContext(ctx, qMetricsTrg); err != nil {
		return err
	}
	return nil
}

// Close closes database connection
func (dbs *DBStorage) Close() {
	logger.Log.Debug().Msg("closing database connection")
	if dbs.db == nil {
		return
	}
	if err := dbs.db.Close(); err != nil {
		logger.Log.Warn().Err(err).Msg("closing database error")
	}
}
