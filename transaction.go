package toki

import (
	"context"
	"database/sql"
	"fmt"
)

// Transaction represents a database transaction
type Transaction struct {
	tx   *sql.Tx
	done bool
}

// TransactionOptions represents options for starting a new transaction
type TransactionOptions struct {
	Isolation sql.IsolationLevel
	ReadOnly  bool
}

// Begin starts a new transaction
func Begin(db *sql.DB) (*Transaction, error) {
	return BeginTx(context.Background(), db, nil)
}

// BeginTx starts a new transaction with options
func BeginTx(ctx context.Context, db *sql.DB, opts *TransactionOptions) (*Transaction, error) {
	var txOpts *sql.TxOptions
	if opts != nil {
		txOpts = &sql.TxOptions{
			Isolation: opts.Isolation,
			ReadOnly:  opts.ReadOnly,
		}
	}

	tx, err := db.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Transaction{tx: tx}, nil
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	if t.done {
		return fmt.Errorf("transaction already committed")
	}

	if err := t.tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	t.done = true
	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	if t.done {
		return fmt.Errorf("transaction already rolled back")
	}

	if err := t.tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	t.done = true
	return nil
}
