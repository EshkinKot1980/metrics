// Модуль storage реализует хранение данных на сервере.
package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	createTablesQuery = `
		CREATE TABLE IF NOT EXISTS counters(
			id VARCHAR(32) PRIMARY KEY,
			delta BIGINT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS gauges(
			id VARCHAR(32) PRIMARY KEY,
			value DOUBLE PRECISION NOT NULL
		);`
	upsertCounterQuery = `
		INSERT INTO counters (id, delta)
			VALUES($1, $2)
		ON CONFLICT (id)
		DO UPDATE SET
			delta = EXCLUDED.delta + counters.delta
		RETURNING delta`
	upsertGaugeQuery = `
		INSERT INTO gauges (id, value)
			VALUES($1, $2)
		ON CONFLICT (id)
		DO UPDATE SET
			value = $2`
	selectCounterQuery = `SELECT delta FROM counters WHERE id = $1`
	selectGaugeQuery   = `SELECT value FROM gauges WHERE id = $1`
)

// Хранилище метрик на основе СУБД Postgres, реализует интерфейс Storage.
type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database is not reachable: %w", err)
	}

	_, err = db.Exec(createTablesQuery)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &DBStorage{db: db}, nil
}

func (s *DBStorage) PutCounter(c Counter) (Counter, error) {
	err := s.db.QueryRow(upsertCounterQuery, c.Name, c.Value).Scan(&c.Value)
	return c, err
}

func (s *DBStorage) PutGauge(g Gauge) error {
	_, err := s.db.Exec(upsertGaugeQuery, g.Name, g.Value)
	return err
}

func (s *DBStorage) GetCounter(c Counter) (Counter, error) {
	err := s.db.QueryRow(selectCounterQuery, c.Name).Scan(&c.Value)
	if err == sql.ErrNoRows {
		err = ErrCounterNotFound
	}
	return c, err
}

func (s *DBStorage) GetGauge(g Gauge) (Gauge, error) {
	err := s.db.QueryRow(selectGaugeQuery, g.Name).Scan(&g.Value)
	if err == sql.ErrNoRows {
		err = ErrGaugeNotFound
	}
	return g, err
}

func (s *DBStorage) PutMetrics(ctx context.Context, counters []Counter, gauges []Gauge) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, c := range counters {
		_, err = s.db.Exec(upsertCounterQuery, c.Name, c.Value)
		if err != nil {
			return err
		}
	}

	for _, g := range gauges {
		_, err = s.db.Exec(upsertGaugeQuery, g.Name, g.Value)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *DBStorage) Halt() {
	s.db.Close()
}

func (s *DBStorage) Ping() bool {
	err := s.db.Ping()
	return err == nil
}
