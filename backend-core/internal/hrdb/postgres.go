package hrdb

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresClient is the production HR DB adapter introduced by ADR-0008
// (sprint claude/work_260513-m). It backs the HRDBClient interface used by
// the auth_signup handler with a row in hrdb.persons; ETL keeps that table
// in sync with the upstream HR system on a daily cadence (ADR-0008 §4.1).
//
// MockClient (mock.go) stays the in-memory dev/test fallback; both share
// the (string, string, string, error) Lookup signature consumed by
// internal/httpapi/router.go::HRDBClient.
type PostgresClient struct {
	Pool *pgxpool.Pool
	// EmailFallbackDomain backs the COALESCE in §4.2 of ADR-0008 — when
	// hrdb.persons.email is NULL, build "<system_id>@<EmailFallbackDomain>".
	// Set from the DEVHUB_HR_EMAIL_FALLBACK_DOMAIN env var; "example.com"
	// is the documented default.
	EmailFallbackDomain string
}

func (c *PostgresClient) Lookup(ctx context.Context, systemID, employeeID, name string) (string, string, string, error) {
	if c.Pool == nil {
		return "", "", "", errors.New("hrdb.PostgresClient: Pool is not configured")
	}
	domain := c.EmailFallbackDomain
	if domain == "" {
		domain = "example.com"
	}

	const q = `
SELECT COALESCE(email, system_id || '@' || $4) AS email,
       system_id,
       department_name
FROM hrdb.persons
WHERE lower(system_id) = lower($1)
  AND employee_id = $2
  AND lower(name) = lower($3)
LIMIT 1
`
	var email, sysID, dept string
	err := c.Pool.QueryRow(ctx, q, systemID, employeeID, name, domain).Scan(&email, &sysID, &dept)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", "", ErrPersonNotFound
		}
		return "", "", "", err
	}
	return email, sysID, dept, nil
}
