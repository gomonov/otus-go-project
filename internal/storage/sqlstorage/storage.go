package sqlstorage

import (
	"github.com/gomonov/otus-go-project/internal/storage"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db *sqlx.DB
}

func NewStorage(connectionString string) (*Storage, error) {
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) Subnet() storage.SubnetRepository {
	return &SubnetRepository{db: s.db}
}
