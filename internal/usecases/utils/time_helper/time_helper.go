package timehelper

import (
	"fmt"
	"time"
)

// StringToTime преобразует строку формата YYYY-MM-DD в time.Time.
func StringToTime(str string) (time.Time, error) {
	res, err := time.Parse(time.DateOnly, str)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse error: %v", err)
	}
	return res, nil
}
