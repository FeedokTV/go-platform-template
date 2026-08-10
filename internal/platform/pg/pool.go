package pg

import (
	"context"
	"go-platform-template/internal/platform/config"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, cfg config.DatabaseConfig, log *slog.Logger) (*pgxpool.Pool, error) {

	pgConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, err
	}

	// MaxConns 10: enough for 1 model and ~5 possible queries. Tune with real load data.
	pgConfig.MaxConns = 10
	pgConfig.MaxConnLifetime = time.Minute * 30
	pgConfig.MaxConnIdleTime = time.Minute * 5

	pool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
