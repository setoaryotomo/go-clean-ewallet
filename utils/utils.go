package utils

import (
	"database/sql"
	"log"
	"strconv"
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

func LogError(svcName, refNo, methodName string, err error, data ...string) {
	str := svcName
	if refNo != "" {
		str += " , referenceNo :: " + refNo
	}
	if methodName != "" {
		str += " , error " + methodName
	}
	if err != nil {
		str += " :: " + err.Error()
	}
	for i, s := range data {
		str += ",\n Data " + strconv.Itoa(i+1) + " ::" + s
	}
	log.Println(str)
}

func LogInfo(svcName, refNo, methodName string, data ...string) {
	str := svcName
	if refNo != "" {
		str += " , referenceNo :: " + refNo
	}
	if methodName != "" {
		str += " , " + methodName
	}
	for i, s := range data {
		str += ",\n Data " + strconv.Itoa(i+1) + " ::" + s
	}
	log.Println(str)
}
