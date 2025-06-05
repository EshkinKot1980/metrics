package pg

import (
	"database/sql"
	"time"

	"github.com/EshkinKot1980/metrics/internal/server/storage"
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
			delta = EXCLUDED.delta + $2
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

type DBStorage struct {
	db *sql.DB
}

func New(conn *sql.DB) (*DBStorage, error) {
	s := &DBStorage{db: conn}
	_, err := s.db.Exec(createTablesQuery)
	return s, err
}

func (s *DBStorage) PutCounter(name string, increment int64) (int64, error) {
	var delta int64
	err := s.db.QueryRow(upsertCounterQuery, name, increment).Scan(&delta)
	return delta, err
}

func (s *DBStorage) PutGauge(name string, value float64) error {
	_, err := s.db.Exec(upsertGaugeQuery, name, value)
	return err
}

func (s *DBStorage) GetCounter(name string) (int64, error) {
	var delta int64
	err := s.db.QueryRow(selectCounterQuery, name).Scan(&delta)
	if err == sql.ErrNoRows {
		err = storage.ErrCounterNotFound
	}
	return delta, err
}

func (s *DBStorage) GetGauge(name string) (float64, error) {
	var value float64
	err := s.db.QueryRow(selectGaugeQuery, name).Scan(&value)
	if err == sql.ErrNoRows {
		err = storage.ErrGaugeNotFound
	}
	return value, err
}

func (s *DBStorage) Halt() {
	<-time.After(time.Duration(1) * time.Second)
}
