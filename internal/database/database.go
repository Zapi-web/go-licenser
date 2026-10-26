package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zapi-web/go-licenser/internal/domain"
)

type Config struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifeTime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectionTimeout time.Duration
}

type PGDB struct {
	psql *pgxpool.Pool
}

func New(ctx context.Context, connString string, cfg Config) (*PGDB, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = cfg.MaxConnLifeTime
	config.MaxConnIdleTime = cfg.MaxConnIdleTime
	config.HealthCheckPeriod = cfg.HealthCheckPeriod
	config.ConnConfig.ConnectTimeout = cfg.ConnectionTimeout

	config.ConnConfig.RuntimeParams = map[string]string{
		"application_name": "go-licenser",
		"timezone":         "UTC",
	}

	db, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgreSQL: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PGDB{psql: db}, nil
}

func (p *PGDB) Set(ctx context.Context, license *domain.License) error {
	query := `
		INSERT INTO licenses (status ,license_key, max_activations, hwid, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := p.psql.QueryRow(ctx, query,
		license.Status,
		license.LicenseKey,
		license.MaxActivations,
		license.HWID,
		license.ExpiresAt,
	).Scan(&license.ID, &license.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert license to database: %w", err)
	}

	return nil
}

func (p *PGDB) Get(ctx context.Context, licenseKey string) (*domain.License, error) {
	query := `
		SELECT id, license_key, status, hwid, max_activations, created_at,expires_at, activated_at
		FROM licenses
		WHERE license_key = $1
	`

	var lic domain.License

	err := p.psql.QueryRow(ctx, query, licenseKey).Scan(
		&lic.ID,
		&lic.LicenseKey,
		&lic.Status,
		&lic.HWID,
		&lic.MaxActivations,
		&lic.CreatedAt,
		&lic.ExpiresAt,
		&lic.ActivatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get license for %s: %w", licenseKey, err)
	}

	return &lic, nil
}

func (p *PGDB) BindHWID(ctx context.Context, licenseID string, hwid string) error {
	query := `
		UPDATE licenses
		SET hwid = $1, activated_at = NOW()
		WHERE id = $2 AND hwid IS NULL
	`
	res, err := p.psql.Exec(ctx, query, hwid, licenseID)

	if err != nil {
		return fmt.Errorf("failed to bind hwid: %w", err)
	}

	if res.RowsAffected() == 0 {
		return errors.New("cannot bind hwid: license not found or already bounded")
	}

	return nil
}

func (p *PGDB) Close() {
	p.psql.Close()
}
