package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
	"github.com/freepaddler/yap-metrics/pkg/retry"
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

	qUpsertGauge = `
		INSERT INTO metrics (name,type,f_value,updated_ts) VALUES ($1,$2,$3,$4)
			ON CONFLICT (name,type)
			DO UPDATE SET f_value = excluded.f_value;
	`
	qUpsertCounter = `
		INSERT INTO metrics (name,type,i_value,updated_ts) VALUES ($1,$2,$3,$4)
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
		logger.Log().Error().Err(err).Msg("unable to setup database connection")
		return nil, err
	}
	logger.Log().Info().Msg("initialize database")
	if err = dbs.initDB(); err != nil {
		// Error here instead of Fatal to let server work without db to pass tests 10[ab]
		logger.Log().Error().Err(err).Msg("unable to init database")
		return nil, err
	}

	return dbs, nil
}

// SaveMetrics inserts metric in database, updates metric value if metric already exists
func (dbs *DBStorage) SaveMetrics(ctx context.Context, metrics []models.Metrics) {

	if len(metrics) == 0 {
		logger.Log().Warn().Msg("no metrics to save")
	}

	err := retry.WithStrategy(ctx,
		func(ctx context.Context) (err error) {
			ctxDB, ctxDBCancel := context.WithTimeout(ctx, DBTimeout*time.Second)
			defer ctxDBCancel()
			if len(metrics) == 1 {
				logger.Log().Debug().Msg("upsert one metric without transaction")
				m := metrics[0]
				switch m.Type {
				case models.Gauge:
					_, err = dbs.db.ExecContext(ctxDB, qUpsertGauge, m.Name, m.Type, *m.FValue, time.Now())
				case models.Counter:
					_, err = dbs.db.ExecContext(ctxDB, qUpsertCounter, m.Name, m.Type, *m.IValue, time.Now())
				}
				if err != nil {
					logger.Log().Warn().Err(err).Msgf("unable to upsert metric '%+v'", m)
					return
				}
			}
			tx, err := dbs.db.BeginTx(ctxDB, nil)
			if err != nil {
				logger.Log().Warn().Err(err).Msg("unable to begin upsert transaction")
				return
			}
			defer tx.Rollback()

			for _, m := range metrics {
				switch m.Type {
				case models.Gauge:
					_, err = tx.ExecContext(ctxDB, qUpsertGauge, m.Name, m.Type, *m.FValue, time.Now())
				case models.Counter:
					_, err = tx.ExecContext(ctxDB, qUpsertCounter, m.Name, m.Type, *m.IValue, time.Now())
				}
				if err != nil {
					logger.Log().Warn().Err(err).Msgf("unable to upsert metric '%+v'", m)
					return
				}
			}
			if err = tx.Commit(); err != nil {
				logger.Log().Warn().Err(err).Msg("unable to commit upsert transaction")
				return
			}
			return
		},
		isRetryErr,
		1, 3, 5)
	if err != nil {
		logger.Log().Error().Err(err).Msg("unable to save metrics to db")
		return
	}
	logger.Log().Debug().Msgf("%d metrics saved to db", len(metrics))
}

func (dbs *DBStorage) RestoreStorage(s store.Storage) {
	logger.Log().Debug().Msg("starting storage restore")
	metrics := dbs.getMetrics()
	s.UpdateMetrics(metrics, true)
	logger.Log().Debug().Msg("done storage restore")
}

func (dbs *DBStorage) SaveStorage(s store.Storage) {
	logger.Log().Debug().Msg("saving store to database")
	snap := s.Snapshot(false)
	dbs.SaveMetrics(context.TODO(), snap)
}

func (dbs *DBStorage) Ping() error {
	err := retry.WithStrategy(context.TODO(),
		func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, DBTimeout*time.Second)
			defer cancel()
			return dbs.db.PingContext(ctx)
		},
		isRetryErr,
		1, 3, 5)
	return err
}

// getMetrics selects all metrics from database
func (dbs *DBStorage) getMetrics() []models.Metrics {
	var res []models.Metrics
	err := retry.WithStrategy(context.TODO(),
		func(ctx context.Context) error {
			ctxDB, ctxDBCancel := context.WithTimeout(ctx, DBTimeout*time.Second)
			defer ctxDBCancel()
			rows, err := dbs.db.QueryContext(ctxDB, `SELECT name, type, f_value, i_value FROM metrics`)
			if err != nil {
				logger.Log().Error().Err(err).Msg("unable to get all metrics from db")
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var (
					m      models.Metrics
					iValue sql.NullInt64
					fValue sql.NullFloat64
				)
				if err = rows.Scan(&m.Name, &m.Type, &fValue, &iValue); err != nil {
					logger.Log().Warn().Err(err).Msg("unable parse metric from db")
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
				logger.Log().Warn().Err(err).Msg("error while getting metrics from db")
			}
			return err
		},
		isRetryErr,
		1, 3, 5)
	if err != nil {
		return nil
	}
	return res
}

// initDB creates necessary database entities: tables, indexes, etc...
func (dbs *DBStorage) initDB() (err error) {
	for _, q := range []string{qMetricsTbl, qMetricsIdx} {
		logger.Log().Debug().Msgf("run init db script %s", q)
		err = retry.WithStrategy(context.TODO(),
			func(ctx context.Context) error {
				ctxDB, ctxDBCancel := context.WithTimeout(ctx, DBTimeout*time.Second)
				defer ctxDBCancel()
				_, err := dbs.db.ExecContext(ctxDB, q)
				return err
			},
			isRetryErr,
			1, 3, 5)
	}
	return
}

// Close closes database connection
func (dbs *DBStorage) Close() {
	logger.Log().Debug().Msg("closing database connection")
	if dbs.db == nil {
		return
	}
	if err := dbs.db.Close(); err != nil {
		logger.Log().Warn().Err(err).Msg("closing database error")
	}
}

// isRetryErr returns true, when error is retryable regarding postgres requests
func isRetryErr(err error) bool {
	if retry.IsNetErr(err) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgerrcode.IsConnectionException(pgErr.Code) {
		return true
	}
	if pgerrcode.IsOperatorIntervention(pgErr.Code) {
		return true
	}
	return false
}
