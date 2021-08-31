package sqltype

import (
	"database/sql/driver"
	"strconv"
	"strings"
)

// 以,分割的字符串数组
type StringArrayString []string

func (j *StringArrayString) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	v := strings.Split(string(bytes), ",")
	*j = v
	return nil
}

func (j StringArrayString) Value() (driver.Value, error) {
	return strings.Join(j, ","), nil
}

// 以,分割的整数数组
type StringArrayInt []int64

func (j *StringArrayInt) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	var res []int64
	v := strings.Split(string(bytes), ",")
	for i := range v {
		val, _ := strconv.ParseInt(v[i], 10, 64)
		res = append(res, val)
	}
	*j = res
	return nil
}

func (j StringArrayInt) Value() (driver.Value, error) {
	var res []string
	for i := range j {
		str := strconv.FormatInt(j[i], 10)
		res = append(res, str)
	}
	return strings.Join(res, ","), nil
}
