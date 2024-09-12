package services

import (
	"github.com/jmoiron/sqlx"
)

type S struct {
	db *sqlx.DB
}

type TransactionFunc func(*sqlx.Tx) error

func New(db *sqlx.DB) *S {
	return &S{db: db}
}

func (s *S) WithTransaction(fn func(tx *sqlx.Tx) error) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(tx)
	return err
}
