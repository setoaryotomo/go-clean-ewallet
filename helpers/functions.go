package helpers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sample/models"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
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
