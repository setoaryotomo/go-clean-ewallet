package utils

import (
	"database/sql"
)

// DBTransaction handles database transactions
func DBTransaction(db *sql.DB, txFunc func(*sql.Tx) error) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Rollback Panic
		} else if err != nil {
			tx.Rollback() // err is not nil
		} else {
			err = tx.Commit() // err is nil
		}
	}()
	err = txFunc(tx)
	return err
}

// TransactionError untuk error kustom dalam transaksi
type TransactionError struct {
	Code    string
	Message string
}

func (e *TransactionError) Error() string {
	return e.Message
}
