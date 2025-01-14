package services

import (
	"github.com/devs-group/sloth/backend/database"
	"github.com/jmoiron/sqlx"
)

type S struct {
	dbService database.IDatabaseService
}

type TransactionFunc func(*sqlx.Tx) error

func New(db database.IDatabaseService) *S {
	return &S{dbService: db}
}

func (s *S) WithTransaction(fn func(tx *sqlx.Tx) error) error {
	tx, err := s.dbService.GetConn().Beginx()
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
