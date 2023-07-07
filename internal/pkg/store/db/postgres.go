package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

func New(uri string) *sql.DB {
	db, err := sql.Open("pgx", uri)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("unable to connect to database")
	}
	return db
}
