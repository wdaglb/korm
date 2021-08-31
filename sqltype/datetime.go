package sqltype

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "2006-01-02 15:04:05"
type DateTime time.Time

func (t *DateTime) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		return nil
	}
	var err error
	timeStr := strings.Trim(str, "\"")
	re, err := regexp.Compile(`^\d+$`)
	if err != nil {
		return err
	}
	if re.Match([]byte(timeStr)) {
		val, _ := strconv.ParseInt(timeStr, 0, 64)
		*t = DateTime(time.Unix(val, 0))
		return err
	}

	t1, err := time.Parse(dateFormat, timeStr)
	*t = DateTime(t1)
	return err
}

func (t DateTime) MarshalJSON() ([]byte, error) {
	now := time.Time(t)
	timestamp := now.Unix()
	var formatted string
	if timestamp == 0 {
		formatted = "null"
	} else {
		formatted = fmt.Sprintf("\"%v\"", now.Format(dateFormat))
	}
	return []byte(formatted), nil
}

func (t *DateTime) Scan(value interface{}) error {
	//s := value.(time.Time)
	switch ty := value.(type) {
	case time.Time:
		*t = DateTime(ty)
	case int64:
		now := time.Unix(ty, 0)
		*t = DateTime(now)
	case []byte:
		now := fmt.Sprintf("%v", string(ty))
		val, _ := strconv.ParseInt(now, 0, 64)
		*t = DateTime(time.Unix(val, 0))
	case string:
		now := fmt.Sprintf("%v", ty)
		val, _ := strconv.ParseInt(now, 0, 64)
		*t = DateTime(time.Unix(val, 0))
	default:
		return errors.New("time format failed")
	}
	return nil
}

func (t DateTime) Value() (driver.Value, error) {
	now := time.Time(t)
	if now.IsZero() {
		return int64(0), nil
	}
	return now.Format(dateFormat), nil
}
