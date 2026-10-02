package api

import (
	"context"
	"embed"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuelsih/golib/httpx"
	"github.com/samuelsih/golib/oas"
	"github.com/samuelsih/golib/sqlmigration"
	"github.com/samuelsih/golib/sqlmigration/pgx"
)

//go:embed migrations/*
var migrationsDir embed.FS

type Config struct {
	DBConnURL         string        `env:"DATABASE_URL,required"`
	DBMaxOpenConn     int32         `env:"DATABASE_MAX_OPEN_CONN"     envDefault:"20"`
	DBMaxLifetimeConn time.Duration `env:"DATABASE_MAX_CONN_LIFETIME" envDefault:"5m"`

	ServerHost                string        `env:"SERVER_HOST"                  envDefault:"127.0.0.1"`
	ServerPort                int           `env:"SERVER_PORT"                  envDefault:"6000"`
	ServerRequestTimeout      time.Duration `env:"SERVER_REQUEST_TIMEOUT"       envDefault:"5s"`
	ServerReadHeaderTimeout   time.Duration `env:"SERVER_READ_HEADER_TIMEOUT"   envDefault:"5s"`
	ServerShutdownTimeout     time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT"      envDefault:"15s"`
	ServerShutdownHardTimeout time.Duration `env:"SERVER_SHUTDOWN_HARD_TIMEOUT" envDefault:"3s"`

	SessionTTL          time.Duration `env:"SESSION_TTL"           envDefault:"720h"`
	SessionCookieName   string        `env:"SESSION_COOKIE_NAME"   envDefault:"nibiru_session"`
	SessionCookieSecure bool          `env:"SESSION_COOKIE_SECURE" envDefault:"true"`
}

func (c Config) HostPort() string {
	return net.JoinHostPort(c.ServerHost, strconv.Itoa(c.ServerPort))
}

func (c Config) ConnectDB(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(c.DBConnURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}

	cfg.MaxConns = c.DBMaxOpenConn
	cfg.MaxConnLifetime = c.DBMaxLifetimeConn

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot create pool: %w", err)
	}

	var pingErr error
	for attempt := 1; attempt <= 3; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		pingErr = pool.Ping(pingCtx)
		cancel()

		if pingErr == nil {
			return pool, nil
		}

		if attempt == 3 {
			break
		}

		timer := time.NewTimer(5 * time.Second)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			pool.Close()
			return nil, ctx.Err()
		}
	}

	pool.Close()
	return nil, fmt.Errorf("cannot connect to database instance: %w", pingErr)
}

func (c Config) MigrationUp(ctx context.Context, db *pgxpool.Pool) error {
	adapter := pgx.New(db)
	return sqlmigration.Up(ctx, adapter, migrationsDir, "migrations")
}

func (c Config) Webserver() *oas.APIServer {
	server := oas.NewServer(httpx.NewRouter(), oas.ServerConfig{
		Title:       "Nibiru API",
		Description: "Nibiru API Documentation",
		License:     null.ValueFrom(oas.LicenseMIT),
		SecuritySchemes: map[string]oas.RefT[oas.SecurityScheme]{
			"cookieAuth": {
				Value: &oas.SecurityScheme{
					Type:        oas.SecuritySchemeTypeAPIKey,
					In:          null.ValueFrom(oas.SecuritySchemeInCookie),
					Name:        null.StringFrom(c.SessionCookieName),
					Description: null.StringFrom("Session cookie issued on login."),
				},
			},
		},
	})

	server.EnableDocUI("/docs")

	return server
}
