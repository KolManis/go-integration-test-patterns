package testdb

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const templateDB = "bakery_template"

// templateLock — номер advisory-блокировки: пакеты, стартующие одновременно, не создают шаблон дважды.
const templateLock = 424242

// NewDatabase создаёт отдельную БД из шаблона с миграциями и удаляет её по окончании.
// Каждый пакет получает свою чистую базу, поэтому пакеты можно запускать параллельно.
func NewDatabase(tb testing.TB) *pgxpool.Pool {
	tb.Helper()
	ctx := context.Background()
	admin := Connect(tb, DSN())

	ensureTemplate(tb, admin)

	name := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, fmt.Sprintf(`CREATE DATABASE %s TEMPLATE %s`, name, templateDB)); err != nil {
		tb.Fatalf("create database %s: %v", name, err)
	}
	tb.Cleanup(func() {
		if _, err := admin.Exec(context.Background(), fmt.Sprintf(`DROP DATABASE IF EXISTS %s WITH (FORCE)`, name)); err != nil {
			tb.Errorf("drop database %s: %v", name, err)
		}
	})

	pool := Connect(tb, withDatabase(tb, DSN(), name))
	return pool
}

// ensureTemplate один раз создаёт шаблонную БД и применяет к ней миграции.
func ensureTemplate(tb testing.TB, admin *pgxpool.Pool) {
	tb.Helper()
	ctx := context.Background()
	conn, err := admin.Acquire(ctx)
	if err != nil {
		tb.Fatalf("acquire: %v", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, templateLock); err != nil {
		tb.Fatalf("lock template: %v", err)
	}
	defer conn.Exec(ctx, `SELECT pg_advisory_unlock($1)`, templateLock) //nolint:errcheck

	var exists bool
	if err := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, templateDB).Scan(&exists); err != nil {
		tb.Fatalf("check template: %v", err)
	}
	if exists {
		return
	}
	if _, err := conn.Exec(ctx, fmt.Sprintf(`CREATE DATABASE %s`, templateDB)); err != nil {
		tb.Fatalf("create template: %v", err)
	}
	// Миграции — отдельным соединением, которое сразу закрываем:
	// шаблон нельзя копировать, пока к нему есть подключения.
	tmpl, err := pgx.Connect(ctx, withDatabase(tb, DSN(), templateDB))
	if err != nil {
		tb.Fatalf("connect template: %v", err)
	}
	defer tmpl.Close(ctx)
	if _, err := tmpl.Exec(ctx, migrationSQL(tb)); err != nil {
		tb.Fatalf("migrate template: %v", err)
	}
}

func withDatabase(tb testing.TB, dsn, name string) string {
	tb.Helper()
	u, err := url.Parse(dsn)
	if err != nil {
		tb.Fatalf("parse dsn: %v", err)
	}
	u.Path = "/" + name
	return u.String()
}
