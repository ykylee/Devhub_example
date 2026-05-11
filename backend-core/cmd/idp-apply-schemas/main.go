// Package main is a one-off helper that applies infra/idp/sql/*.sql to the
// devhub PostgreSQL database. It exists because psql.exe is not on the dev
// workstation PATH; reusing backend-core's pgx/v5 dependency avoids fetching
// new modules in the corp-restricted Go environment.
//
// Run from repo root:
//
//	go run ./backend-core/cmd/idp-apply-schemas
//	go run ./backend-core/cmd/idp-apply-schemas -dsn "postgres://..." -sql infra/idp/sql/001_create_idp_schemas.sql
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	defaultDSN := "postgres://postgres:postgres@127.0.0.1:5432/devhub?sslmode=disable"
	if v := os.Getenv("DEVHUB_DB_URL"); v != "" {
		defaultDSN = v
	}

	defaultSQL := filepath.Join("infra", "idp", "sql", "001_create_idp_schemas.sql")

	dsn := flag.String("dsn", defaultDSN, "PostgreSQL DSN (or set DEVHUB_DB_URL).")
	sqlPath := flag.String("sql", defaultSQL, "Path to the SQL file to execute.")
	query := flag.String("query", "", "Optional ad-hoc SELECT statement to print row-by-row (skips file execution when set).")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, *dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(context.Background())

	fmt.Printf("Connected: %s\n", redactPassword(*dsn))

	if *query != "" {
		rows, err := conn.Query(ctx, *query)
		if err != nil {
			log.Fatalf("query: %v", err)
		}
		defer rows.Close()
		fields := rows.FieldDescriptions()
		names := make([]string, len(fields))
		for i, f := range fields {
			names[i] = string(f.Name)
		}
		fmt.Println(strings.Join(names, "\t"))
		for rows.Next() {
			vals, err := rows.Values()
			if err != nil {
				log.Fatalf("scan: %v", err)
			}
			parts := make([]string, len(vals))
			for i, v := range vals {
				parts[i] = fmt.Sprintf("%v", v)
			}
			fmt.Println(strings.Join(parts, "\t"))
		}
		if err := rows.Err(); err != nil {
			log.Fatalf("rows err: %v", err)
		}
		return
	}

	body, err := os.ReadFile(*sqlPath)
	if err != nil {
		log.Fatalf("read sql file %s: %v", *sqlPath, err)
	}

	fmt.Printf("Executing %s ...\n", *sqlPath)

	if _, err := conn.Exec(ctx, string(body)); err != nil {
		log.Fatalf("exec sql: %v", err)
	}

	rows, err := conn.Query(ctx, "SELECT schema_name FROM information_schema.schemata WHERE schema_name IN ('hydra','kratos') ORDER BY schema_name")
	if err != nil {
		log.Fatalf("verify query: %v", err)
	}
	defer rows.Close()

	fmt.Println("Schemas present:")
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatalf("scan: %v", err)
		}
		fmt.Printf("  - %s\n", name)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rows err: %v", err)
	}

	fmt.Println("Done.")
}

func redactPassword(dsn string) string {
	if i := strings.Index(dsn, "://"); i >= 0 {
		rest := dsn[i+3:]
		if at := strings.Index(rest, "@"); at >= 0 {
			creds := rest[:at]
			if colon := strings.Index(creds, ":"); colon >= 0 {
				return dsn[:i+3] + creds[:colon] + ":***" + rest[at:]
			}
		}
	}
	return dsn
}
