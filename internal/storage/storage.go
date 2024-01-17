package storage

import "github.com/jmoiron/sqlx"

type DBStorage struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *DBStorage {
	return &DBStorage{db: db}
}
