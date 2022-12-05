package tx

import (
	"context"
	"database/sql"
	"log"
	"os"
)

type TransactionManager[T any] interface {
	ExecTx(ctx context.Context, fn Fn[T], repository Repository) (T, error)
	GetTx() *sql.Tx
}

type transaction[T any] struct {
	opts   *sql.TxOptions
	tx     *sql.Tx
	logger *log.Logger
}

type Fn[T any] func(ctx context.Context, tx TransactionManager[T]) (T, error)

func NewTransaction[T any]() TransactionManager[T] {
	return &transaction[T]{}
}

func newTx[T any](ctx context.Context, db Repository, opts *sql.TxOptions) (transaction[T], error) {
	var tx transaction[T]
	sqlTx, err := db.GetDB().BeginTx(ctx, opts)
	if err != nil {
		return tx, err
	}

	return transaction[T]{
		opts:   opts,
		tx:     sqlTx,
		logger: log.New(os.Stdout, "", 5),
	}, nil
}

func (t transaction[T]) ExecTx(ctx context.Context, fn Fn[T], repository Repository) (T, error) {
	var err error
	var res T
	nTx, err := newTx[T](ctx, repository, nil)
	if err != nil {
		return res, err
	}

	res, err = fn(ctx, nTx)
	if err != nil {
		return res, nTx.checkTransaction(err)
	}

	return res, nTx.checkTransaction(err)
}

func (t transaction[T]) checkTransaction(err error) error {
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

func (t transaction[T]) GetTx() *sql.Tx {
	return t.tx
}

type mockTransactionManager[T any] struct {
}

func NewMockTransaction[T any]() TransactionManager[T] {
	return &mockTransactionManager[T]{}
}

func (m mockTransactionManager[T]) GetTx() *sql.Tx {
	return nil
}

func (m mockTransactionManager[T]) ExecTx(ctx context.Context, fn Fn[T], repository Repository) (T, error) {
	mockTransactionManager := mockTransactionManager[T]{}
	return fn(ctx, mockTransactionManager)
}
