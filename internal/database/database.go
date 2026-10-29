package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (p *PGDB) Create(ctx context.Context, license *domain.License) error {
	query := `
		INSERT INTO licenses (status ,license_key, hwid, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := p.psql.QueryRow(ctx, query,
		license.Status,
		license.LicenseKey,
		license.HWID,
		license.ExpiresAt,
	).Scan(&license.ID, &license.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to insert license: %w", err)
	}

	return nil
}

func (p *PGDB) Get(ctx context.Context, licenseKey string) (*domain.License, error) {
	query := `
		SELECT id, license_key, status, hwid, created_at,expires_at, activated_at
		FROM licenses
		WHERE license_key = $1
	`

	var lic domain.License

	err := p.psql.QueryRow(ctx, query, licenseKey).Scan(
		&lic.ID,
		&lic.LicenseKey,
		&lic.Status,
		&lic.HWID,
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

func (p *PGDB) BindHWID(ctx context.Context, licenseKey string, hwid string) error {
	query := `
		UPDATE licenses
		SET hwid = $1, activated_at = NOW()
		WHERE license_key = $2 AND hwid IS NULL
	`

	res, err := p.psql.Exec(ctx, query, hwid, licenseKey)
	if err != nil {
		return fmt.Errorf("failed to bind hwid: %w", err)
	}

	if res.RowsAffected() > 0 {
		return nil
	}

	var currentHWID *string
	checkQuery := `SELECT hwid FROM licenses WHERE license_key = $1`
	err = p.psql.QueryRow(ctx, checkQuery, licenseKey).Scan(&currentHWID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("failed to verify license state: %w", err)
	}

	switch {
	case currentHWID == nil:
		return domain.ErrConcurrentModification
	case *currentHWID == hwid:
		return nil
	default:
		return domain.ErrAlreadyBound
	}
}

func (p *PGDB) ResetHWID(ctx context.Context, licenseKey string) error {
	query := `
		UPDATE licenses
		SET hwid = NULL, activated_at = NULL
		WHERE license_key = $1
	`

	err := p.performQuery(ctx, query, licenseKey)
	if err != nil {
		return fmt.Errorf("failed to reset hwid: %w", err)
	}

	return nil
}

func (p *PGDB) ExtendExpiry(ctx context.Context, licenseKey string, newExpiresAt time.Time) error {
	query := `
		UPDATE licenses
		SET expires_at = $1
		WHERE license_key = $2
	`

	err := p.performQuery(ctx, query, newExpiresAt, licenseKey)
	if err != nil {
		return fmt.Errorf("failed to extend expiry: %w", err)
	}

	return nil
}

func (p *PGDB) UpdateStatus(ctx context.Context, licenseKey string, newStatus domain.Status) error {
	query := `
		UPDATE licenses
		SET status = $1
		WHERE license_key = $2
	`

	err := p.performQuery(ctx, query, newStatus, licenseKey)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func (p *PGDB) performQuery(ctx context.Context, query string, args ...any) error {
	res, err := p.psql.Exec(ctx, query, args...)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (p *PGDB) Close() {
	p.psql.Close()
}
