package tx

import "database/sql"

type Repository[T any] interface {
	GetDB() *sql.DB
	GetTransactionManager() TransactionManager[T]
}
