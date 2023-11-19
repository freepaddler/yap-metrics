package postgres

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
	"github.com/freepaddler/yap-metrics/pkg/retry"
)

const (
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
)

type PostgresStore struct {
	db        *sql.DB
	dbTimeout time.Duration
	retry     []int
}

func WithTimeout(to time.Duration) func(ps *PostgresStore) {
	return func(ps *PostgresStore) {
		ps.dbTimeout = to
	}
}

func WithRetry(retries ...int) func(ps *PostgresStore) {
	return func(ps *PostgresStore) {
		ps.retry = append(ps.retry, retries...)
	}
}

// NewPostgresStorage connects to database and returns Store in case of success
func NewPostgresStorage(uri string, opts ...func(store *PostgresStore)) (*PostgresStore, error) {
	ps := &PostgresStore{dbTimeout: time.Second}
	for _, o := range opts {
		o(ps)
	}
	var err error
	ps.db, err = sql.Open("pgx", uri)
	if err != nil {
		// Error here instead of Fatal to let server work without db to pass tests 10[ab]
		logger.Log().Error().Err(err).Msg("unable to setup database connection to '%s', uri")
		return nil, err
	}
	logger.Log().Info().Msg("initialize database")
	if err = ps.initDB(); err != nil {
		// Error here instead of Fatal to let server work without db to pass tests 10[ab]
		logger.Log().Error().Err(err).Msg("unable to init database")
		return nil, err
	}
	return ps, nil
}

// initDB creates necessary database entities: tables, indexes, etc...
func (ps *PostgresStore) initDB() (err error) {
	for _, q := range []string{qMetricsTbl, qMetricsIdx} {
		logger.Log().Debug().Msgf("run db init %s", q)
		err = retry.WithStrategy(context.TODO(),
			func(ctx context.Context) error {
				ctxDB, ctxDBCancel := context.WithTimeout(ctx, ps.dbTimeout)
				defer ctxDBCancel()
				_, err := ps.db.ExecContext(ctxDB, q)
				return err
			},
			isRetryErr,
			1, 3, 5)
	}
	return
}

func (ps *PostgresStore) SetGauge(name string, value float64) (res float64) {
	logger.Log().Debug().Msgf("SetGauge: store value %f for gauge %s", value, name)
	err := retry.WithStrategy(context.TODO(),
		func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, ps.dbTimeout)
			defer cancel()
			return ps.db.QueryRowContext(ctx, `
				INSERT INTO metrics (name,type,f_value,updated_ts) VALUES ($1,$2,$3,$4)
				ON CONFLICT (name,type)
					DO UPDATE SET f_value = excluded.f_value
				RETURNING f_value`,
				name, models.Gauge, value, time.Now()).Scan(&res)
		},
		isRetryErr,
		1, 3, 5)
	if err != nil {
		logger.Log().Err(err).Msg("SetGauge: failed")
	}
	return
}

func (ps *PostgresStore) GetGauge(name string) (res float64, found bool) {
	err := retry.WithStrategy(context.TODO(),
		func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, ps.dbTimeout)
			defer cancel()
			return ps.db.QueryRowContext(ctx, `
				SELECT f_value FROM metrics WHERE name=$1 and type=$2`,
				name, models.Gauge).Scan(&res)
		},
		isRetryErr,
		1, 3, 5)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		logger.Log().Err(err).Msg("GetGauge: failed")
	}
	found = true
	return
}
func (ps *PostgresStore) DelGauge(name string) {
	err := retry.WithStrategy(context.TODO(),
		func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, ps.dbTimeout)
			defer cancel()
			_, err := ps.db.ExecContext(ctx, `
				DELETE FROM metrics WHERE name=$1 and type=$2`,
				name, models.Gauge)
			return err
		},
		isRetryErr,
		1, 3, 5)
	if err != nil {
		logger.Log().Err(err).Msg("DelGauge: failed")
	}
}

func (ps *PostgresStore) IncCounter(name string, value int64) (res int64) {
	logger.Log().Debug().Msgf("IncCounter: add increment %d for counter %s", value, name)
	err := retry.WithStrategy(context.TODO(),
		func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, ps.dbTimeout)
			defer cancel()
			return ps.db.QueryRowContext(ctx, `
				INSERT INTO metrics (name,type,i_value,updated_ts) VALUES ($1,$2,$3,$4)
				ON CONFLICT (name,type)
					DO UPDATE SET i_value = excluded.i_value + metrics.i_value
				RETURNING i_value`,
				name, models.Counter, value, time.Now()).Scan(&res)
		},
		isRetryErr,
		1, 3, 5)
	if err != nil {
		logger.Log().Err(err).Msg("IncCounter: failed")
	}
	return
}

func (ps *PostgresStore) GetCounter(name string) (res int64, found bool) {
	err := retry.WithStrategy(context.TODO(),
		func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, ps.dbTimeout)
			defer cancel()
			return ps.db.QueryRowContext(ctx, `
				SELECT i_value FROM metrics WHERE name=$1 and type=$2`,
				name, models.Counter).Scan(&res)
		},
		isRetryErr,
		1, 3, 5)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		logger.Log().Err(err).Msg("GetCounter: failed")
	}
	found = true
	return
}

func (ps *PostgresStore) DelCounter(name string) {
	err := retry.WithStrategy(context.TODO(),
		func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, ps.dbTimeout)
			defer cancel()
			_, err := ps.db.ExecContext(ctx, `
				DELETE FROM metrics WHERE name=$1 and type=$2`,
				name, models.Counter)
			return err
		},
		isRetryErr,
		1, 3, 5)
	if err != nil {
		logger.Log().Err(err).Msg("DelCounter: failed")
	}
}

func (ps *PostgresStore) Snapshot(flush bool) (metrics []models.Metrics) {
	err := retry.WithStrategy(context.TODO(),
		func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, ps.dbTimeout)
			defer cancel()
			tx, err := ps.db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}
			defer tx.Rollback()
			rows, err := tx.QueryContext(ctx, `SELECT name,type,i_value,f_value FROM metrics`)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil
				}
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var m models.Metrics
				if err := rows.Scan(&m.Name, &m.Type, &m.IValue, &m.FValue); err != nil {
					return err
				}
				switch m.Type {
				case models.Gauge:
					m.IValue = nil
				case models.Counter:
					m.FValue = nil
				}
				metrics = append(metrics, m)
			}
			if flush {
				tx.ExecContext(ctx, `TRUNCATE TABLE metrics`)
			}
			return tx.Commit()
		},
		isRetryErr,
		1, 3, 5)
	if err != nil {
		logger.Log().Err(err).Msg("DelCounter: failed")
	}
	return
}

func (ps *PostgresStore) Ping() error {
	err := retry.WithStrategy(context.TODO(),
		func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, ps.dbTimeout)
			defer cancel()
			return ps.db.PingContext(ctx)
		},
		isRetryErr,
		1, 3, 5)
	return err
}

// Close closes database connection
func (ps *PostgresStore) Close() {
	logger.Log().Debug().Msg("closing database connection")
	if ps.db == nil {
		return
	}
	if err := ps.db.Close(); err != nil {
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
