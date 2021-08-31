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

type Timestamp time.Time

func (t *Timestamp) UnmarshalJSON(data []byte) error {
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
		*t = Timestamp(time.Unix(val, 0))
		return err
	}

	t1, err := time.Parse(dateFormat, timeStr)
	*t = Timestamp(t1)
	return err
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
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

func (t *Timestamp) Scan(value interface{}) error {
	//s := value.(time.Time)
	switch ty := value.(type) {
	case time.Time:
		*t = Timestamp(ty)
	case int64:
		now := time.Unix(ty, 0)
		*t = Timestamp(now)
	case []byte:
		now := fmt.Sprintf("%v", string(ty))
		val, _ := strconv.ParseInt(now, 0, 64)
		*t = Timestamp(time.Unix(val, 0))
	default:
		return errors.New("time format failed")
	}
	return nil
}

func (t Timestamp) Value() (driver.Value, error) {
	now := time.Time(t)
	if now.IsZero() {
		return int64(0), nil
	}
	// return now.Format(dateFormat), nil
	return now.Unix(), nil
}
