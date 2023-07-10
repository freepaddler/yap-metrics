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
	DBTimeout   = 30 // database query timeout
	qMetricsTbl = `
		CREATE TABLE IF NOT EXISTS metrics 	(
			id      INT GENERATED ALWAYS AS IDENTITY,
			updated_ts TIMESTAMPTZ NOT NULL DEFAULT current_timestamp(3),
			name    VARCHAR NOT NULL,
			type    VARCHAR NOT NULL,
			f_value DOUBLE PRECISION,
			i_value BIGINT
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
	qUpsertGauge = `
		INSERT INTO metrics (name,type,f_value) VALUES ($1,$2,$3)
			ON CONFLICT (name,type)
			DO UPDATE SET f_value = excluded.f_value;
	`
	qUpsertCounter = `
		INSERT INTO metrics (name,type,i_value) VALUES ($1,$2,$3)
			ON CONFLICT (name,type)
			DO UPDATE SET i_value = excluded.i_value;
	`
)

type DBStorage struct {
	db *sql.DB
}

// New is a DBStorage constructor
func New(uri string) (*DBStorage, error) {
	dbs := new(DBStorage)
	var err error
	dbs.db, err = sql.Open("pgx", uri)
	if err != nil {
		// Error here instead of Fatal to let server work without db to pass tests 10[ab]
		logger.Log.Error().Err(err).Msg("unable to setup database connection")
		return nil, err
	}
	if err = dbs.initDB(); err != nil {
		// Error here instead of Fatal to let server work without db to pass tests 10[ab]
		logger.Log.Error().Err(err).Msg("unable to init database")
		return nil, err
	}
	return dbs, nil
}

// SaveMetrics inserts metric in database, updates metric value if metric already exists
func (dbs *DBStorage) SaveMetrics(metrics []models.Metrics) {
	ctxDB, ctxDBCancel := context.WithTimeout(context.Background(), DBTimeout*time.Second)
	defer ctxDBCancel()
	if len(metrics) == 0 {
		logger.Log.Warn().Msg("no metrics to save")
	}
	var err error
	if len(metrics) == 1 {

		logger.Log.Debug().Msg("upsert one metric without transaction")
		m := metrics[0]
		switch m.Type {
		case models.Gauge:
			_, err = dbs.db.ExecContext(ctxDB, qUpsertGauge, m.Name, m.Type, *m.FValue)
		case models.Counter:
			_, err = dbs.db.ExecContext(ctxDB, qUpsertCounter, m.Name, m.Type, *m.IValue)
		}
		if err != nil {
			logger.Log.Error().Err(err).Msgf("unable to upsert metric '%+v'", m)
			return
		}
	}
	tx, err := dbs.db.BeginTx(ctxDB, nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("unable to begin upsert transaction")
		return
	}
	defer tx.Rollback()

	for _, m := range metrics {
		switch m.Type {
		case models.Gauge:
			_, err = tx.ExecContext(ctxDB, qUpsertGauge, m.Name, m.Type, *m.FValue)
		case models.Counter:
			_, err = tx.ExecContext(ctxDB, qUpsertCounter, m.Name, m.Type, *m.IValue)
		}
		if err != nil {
			logger.Log.Error().Err(err).Msgf("unable to upsert metric '%+v'", m)
			return
		}
	}
	if err = tx.Commit(); err != nil {
		logger.Log.Error().Err(err).Msg("unable to commit upsert transaction")
		return
	}
	logger.Log.Debug().Msgf("%d metrics saved to db", len(metrics))
}

func (dbs *DBStorage) RestoreStorage(s store.Storage) {
	logger.Log.Debug().Msg("starting storage restore")
	metrics := dbs.getMetrics()
	s.UpdateMetrics(metrics, true)
	//for _, m := range metrics {
	//	switch m.Type {
	//	case models.Gauge:
	//		s.SetGauge(m.Name, *m.FValue)
	//	case models.Counter:
	//		s.DelCounter(m.Name)
	//		s.IncCounter(m.Name, *m.IValue)
	//	}
	//}
	logger.Log.Debug().Msg("done storage restore")
}

func (dbs *DBStorage) SaveStorage(s store.Storage) {
	logger.Log.Debug().Msg("saving store to database")
	snap := s.Snapshot(false)
	dbs.SaveMetrics(snap)
}

func (dbs *DBStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout*time.Second)
	defer cancel()
	return dbs.db.PingContext(ctx)
}

// getMetrics selects all metrics from database
func (dbs *DBStorage) getMetrics() []models.Metrics {
	ctxDB, ctxDBCancel := context.WithTimeout(context.Background(), DBTimeout*time.Second)
	defer ctxDBCancel()
	var res []models.Metrics
	rows, err := dbs.db.QueryContext(ctxDB, `SELECT name, type, f_value, i_value FROM metrics`)
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

// initDB creates necessary database entities: tables, indexes, etc...
func (dbs *DBStorage) initDB() error {
	ctxDB, ctxDBCancel := context.WithTimeout(context.Background(), DBTimeout*time.Second)
	defer ctxDBCancel()
	if _, err := dbs.db.ExecContext(ctxDB, qMetricsTbl); err != nil {
		return err
	}
	if _, err := dbs.db.ExecContext(ctxDB, qMetricsIdx); err != nil {
		return err
	}
	if _, err := dbs.db.ExecContext(ctxDB, qFuncSetUpdatedTS); err != nil {
		return err
	}
	if _, err := dbs.db.ExecContext(ctxDB, qMetricsTrg); err != nil {
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
