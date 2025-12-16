package utils

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
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

// GenerateReferenceNo generates a unique 16-character reference number
// Format: YYYYMMDDHHMMSS + 2 random uppercase hex characters (total 16 chars)
// Example: 2025121614302501
func GenerateReferenceNo() string {
	timestamp := time.Now().Format("20060102150405") // 14 characters
	randomBytes := make([]byte, 1)                   // 1 byte = 2 hex characters
	rand.Read(randomBytes)
	randomHex := strings.ToUpper(hex.EncodeToString(randomBytes))
	return fmt.Sprintf("%s%s", timestamp, randomHex) // 14 + 2 = 16 characters
}

// GenerateReferenceNoWithPrefix generates a reference number with custom prefix
// Format: PREFIX + YYYYMMDDHHMMSS + 6 random hex characters
func GenerateReferenceNoWithPrefix(prefix string) string {
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 3)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s%s%s", prefix, timestamp, randomHex)
}

// GenerateShortReferenceNo generates a shorter reference number
// Format: REF + YYMMDDHHMMSS + 4 random hex
// Example: REF251216143025A1B2
func GenerateShortReferenceNo() string {
	timestamp := time.Now().Format("060102150405")
	// randomBytes := make([]byte, 2) // 2 bytes = 4 hex characters
	randomBytes := make([]byte, 1)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("RF%s%s", timestamp, randomHex)
}
