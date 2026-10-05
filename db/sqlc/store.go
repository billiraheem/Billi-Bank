package db

import (
	"context"
	"database/sql"
	"fmt"
)

// defines all functions to execute db queries and transactions
type Store interface {
	Querier;
	TransferTx(ctx context.Context, args TransferTxParams) (TransferTxResult, error);
	DepositTx(ctx context.Context, args DepositTxParams) (DepositTxResult, error);
	CreateUserTx(ctx context.Context, args CreateUserTxParams) (CreateUserTxResult, error)
	VerifyEmailTx(ctx context.Context, args VerifyEmailTxParams) (VerifyEmailTxResult, error)
}

// provides all functions to execute Db queries and transaction
type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a () within a DB transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}

		return err
	}

	return tx.Commit()
}