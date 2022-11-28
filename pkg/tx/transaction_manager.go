package tx

import (
	"context"
	"database/sql"
	"log"
	"os"
)

type Transaction[T any] struct {
	opts   *sql.TxOptions
	tx     *sql.Tx
	logger *log.Logger
}

type Fn[T any] func(ctx context.Context, tx Transaction[T]) (T, error)

func NewTx[T any](ctx context.Context, db Repository, opts *sql.TxOptions) (Transaction[T], error) {
	var tx Transaction[T]
	sqlTx, err := db.GetDB().BeginTx(ctx, opts)
	if err != nil {
		return tx, err
	}

	return Transaction[T]{
		opts:   opts,
		tx:     sqlTx,
		logger: log.New(os.Stdout, "", 5),
	}, nil
}

func ExecTx[T any](ctx context.Context, fn Fn[T], repository Repository) (T, error) {
	var err error
	var res T
	newTx, err := NewTx[T](ctx, repository, nil)
	if err != nil {
		return res, err
	}

	res, err = fn(ctx, newTx)
	if err != nil {
		return res, newTx.checkTransaction(err)
	}

	return res, newTx.checkTransaction(err)
}

func (t Transaction[T]) checkTransaction(err error) error {
	if err != nil {
		txErr := t.tx.Rollback()
		if txErr != nil {
			t.logger.Printf("transaction rollback error: %w", txErr)
		}
		return err
	}
	err = t.tx.Commit()
	if err != nil {
		t.logger.Printf("transaction commit error: %w", err)
	}
	return err
}

func (t Transaction[T]) GetTx() *sql.Tx {
	return t.tx
}
