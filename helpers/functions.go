package helpers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"sample/models"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
)

func InArray(v interface{}, in interface{}) (ok bool, i int) {
	val := reflect.Indirect(reflect.ValueOf(in))
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for ; i < val.Len(); i++ {
			if ok = v == val.Index(i).Interface(); ok {
				return
			}
		}
	}
	return
}

func BindValidateStruct(ctx echo.Context, i interface{}) error {
	if err := ctx.Bind(i); err != nil {
		return err
	}

	if err := ctx.Validate(i); err != nil {
		return err
	}
	return nil
}

func ResponseJSON(success bool, code string, msg string, result interface{}) models.Response {
	tm := time.Now()
	response := models.Response{
		Success:          success,
		StatusCode:       code,
		Result:           result,
		Message:          msg,
		ResponseDatetime: tm,
	}

	return response
}

func LOG(title string, i interface{}, end bool) {
	log, _ := json.Marshal(i)
	logString := string(log)
	fmt.Println(title, logString)
	if end {
		fmt.Println(title, strings.Repeat("_", 50), "END LOG", strings.Repeat("_", 50))
	}
}

// ReplaceSQL ...
func ReplaceSQL(old, searchPattern string) string {
	tmpCount := strings.Count(old, searchPattern)
	for m := 1; m <= tmpCount; m++ {
		old = strings.Replace(old, searchPattern, "$"+strconv.Itoa(m), 1)
	}
	return old
}

func WeekRange(year, week int) (start, end time.Time) {
	start = WeekStart(year, week)
	end = start.AddDate(0, 0, 6)
	return
}

func WeekStart(year, week int) time.Time {
	// Start from the middle of the year:
	t := time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC)

	// Roll back to Monday:
	if wd := t.Weekday(); wd == time.Sunday {
		t = t.AddDate(0, 0, -6)
	} else {
		t = t.AddDate(0, 0, -int(wd)+1)
	}

	// Difference in weeks:
	_, w := t.ISOWeek()
	t = t.AddDate(0, 0, (week-w)*7)

	return t
}

// Helper functions
func IsNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func HashPIN(pin string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pin), 4)
	return string(bytes), err
}

func CheckPINHash(pin, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin))
	return err == nil
}

func GenerateAccountNumber() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	timestamp := time.Now().UnixNano() / 1000000
	random := rand.Intn(10000)
	number := fmt.Sprintf("%d%04d", timestamp, random)

	if len(number) > 10 {
		number = number[len(number)-10:]
	} else if len(number) < 10 {
		number = fmt.Sprintf("%010s", number)
	}

	return number
}

func GenerateResetToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func Contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || Contains(s[1:], substr)))
}

// transactionDto helper untuk mapping rows ke struct
func TransactionDto(rows *sql.Rows) ([]models.Transaction, error) {
	var result []models.Transaction

	for rows.Next() {
		var val models.Transaction
		var sourceNumber, beneficiaryNumber sql.NullString

		err := rows.Scan(
			&val.ID,
			&val.AccountID,
			&val.AccountNumber,
			&val.AccountName,
			&sourceNumber,
			&beneficiaryNumber,
			&val.TransactionType,
			&val.Amount,
			&val.TransactionTime,
			&val.CreatedAt,
		)
		if err != nil {
			return result, err
		}

		val.SourceNumber = sourceNumber.String
		val.BeneficiaryNumber = beneficiaryNumber.String

		result = append(result, val)
	}

	return result, nil
}

// Helper functions
func NullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
