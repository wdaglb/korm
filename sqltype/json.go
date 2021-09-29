package sqltype

import (
	"bytes"
	"database/sql/driver"
	"errors"
)

type KJson []byte

func (t *KJson) UnmarshalJSON(data []byte) error {
	if t == nil {
		return errors.New("null point exception")
	}
	*t = append((*t)[0:0], data...)
	return nil
}

func (t KJson) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("null"), nil
	}
	return t, nil
}

func (t *KJson) Scan(value interface{}) error {
	if value == nil {
		*t = nil
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return nil
	}
	*t = append((*t)[0:0], v...)
	return nil
}

func (t KJson) Value() (driver.Value, error) {
	if t.IsNull() {
		return nil, nil
	}
	return string(t), nil
}

func (t KJson) IsNull() bool {
	return len(t) == 0 || string(t) == "null"
}

func (t KJson) Equals(val KJson) bool {
	return bytes.Equal(t, val)
}
