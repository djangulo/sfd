package models

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// NullString cousin of sql.NullString that implements json's Marshaler and
// Unmarshaler interfaces.
type NullString struct {
	*sql.NullString
}

func NewNullString(str string) NullString {
	if str == "" {
		return NullString{NullString: &sql.NullString{Valid: false, String: ""}}
	}
	return NullString{NullString: &sql.NullString{Valid: true, String: str}}
}

// Value implements the driver.Valuer interface.
func (ns NullString) Value() (driver.Value, error) {
	if ns.NullString == nil || !ns.NullString.Valid {
		return nil, nil
	}
	// Delegate to the nullstring's Value function
	return ns.NullString.Value()
}

// Scan implements the sql.Scanner interface.
func (ns *NullString) Scan(src interface{}) error {
	if ns.NullString == nil {
		ns.NullString = &sql.NullString{}
	}
	if src == nil {
		ns.NullString.String, ns.NullString.Valid = "", false
		return nil
	}

	// Delegate to NullString Scan function
	return ns.NullString.Scan(src)
}

// MarshalJSON marshals the NullString as null or the nested string.
func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.NullString == nil || !ns.NullString.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(ns.NullString.String)
}

// UnmarshalJSON unmarshals a NullString.
func (ns *NullString) UnmarshalJSON(b []byte) error {
	if ns.NullString == nil {
		ns.NullString = &sql.NullString{}
	}
	if bytes.Equal(b, []byte("null")) {
		ns.NullString.String, ns.NullString.Valid = "", false
		return nil
	}

	if err := json.Unmarshal(b, &ns.NullString.String); err != nil {
		return err
	}

	ns.NullString.Valid = true

	return nil
}

// NullInt64 cousin of sql.NullInt64 that implements json's Marshaler and
// Unmarshaler interfaces.
type NullInt64 struct {
	*sql.NullInt64
}

func NewNullInt64(n int) NullInt64 {
	if n == 0 {
		return NullInt64{NullInt64: &sql.NullInt64{Valid: false, Int64: 0}}
	}
	return NullInt64{NullInt64: &sql.NullInt64{Valid: true, Int64: int64(n)}}
}

// Value implements the driver.Valuer interface.
func (ns NullInt64) Value() (driver.Value, error) {
	if ns.NullInt64 == nil || !ns.NullInt64.Valid {
		return nil, nil
	}
	// Delegate to the NullInt64's Value function
	return ns.NullInt64.Value()
}

// Scan implements the sql.Scanner interface.
func (ns *NullInt64) Scan(src interface{}) error {
	if ns.NullInt64 == nil {
		ns.NullInt64 = &sql.NullInt64{}
	}
	if src == nil {
		ns.NullInt64.Int64, ns.NullInt64.Valid = 0, false
		return nil
	}

	// Delegate to NullInt64 Scan function
	return ns.NullInt64.Scan(src)
}

// MarshalJSON marshals the NullInt64 as null or the nested int64.
func (ns NullInt64) MarshalJSON() ([]byte, error) {
	if ns.NullInt64 == nil || !ns.NullInt64.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(ns.NullInt64.Int64)
}

// UnmarshalJSON unmarshals a NullInt64
func (ns *NullInt64) UnmarshalJSON(b []byte) error {
	if ns.NullInt64 == nil {
		ns.NullInt64 = &sql.NullInt64{}
	}
	if bytes.Equal(b, []byte("null")) {
		ns.NullInt64.Int64, ns.NullInt64.Valid = 0, false
		return nil
	}

	if err := json.Unmarshal(b, &ns.NullInt64.Int64); err != nil {
		return err
	}

	ns.NullInt64.Valid = true

	return nil
}

// NullTime cousin of sql.NullTime that implements json's Marshaler and
// Unmarshaler interfaces.
type NullTime struct {
	*sql.NullTime
}

func NewNullTime(t time.Time) NullTime {
	if t.IsZero() {
		return NullTime{NullTime: &sql.NullTime{Valid: false, Time: time.Time{}}}
	}
	return NullTime{NullTime: &sql.NullTime{Valid: true, Time: t}}
}

// Scan implements the Scanner interface.
func (n *NullTime) Scan(value interface{}) error {
	if n.NullTime == nil {
		n.NullTime = &sql.NullTime{}
	}
	if value == nil {
		n.NullTime.Time, n.NullTime.Valid = time.Time{}, false
		return nil
	}
	n.Valid = true
	return n.NullTime.Scan(value)
}

// Value implements the driver Valuer interface.
func (n NullTime) Value() (driver.Value, error) {
	if n.NullTime == nil || !n.NullTime.Valid {
		return nil, nil
	}
	return n.NullTime.Time, nil
}

// MarshalJSON marshals the NullInt64 as null or the nested int64.
func (ns NullTime) MarshalJSON() ([]byte, error) {
	if ns.NullTime == nil || !ns.NullTime.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(ns.NullTime.Time)
}

// UnmarshalJSON unmarshals a NullInt64
func (ns *NullTime) UnmarshalJSON(b []byte) error {
	if ns.NullTime == nil {
		ns.NullTime = &sql.NullTime{}
	}
	if bytes.Equal(b, []byte("null")) {
		ns.NullTime.Time, ns.NullTime.Valid = time.Time{}, false
		return nil
	}

	if err := json.Unmarshal(b, &ns.NullTime.Time); err != nil {
		return err
	}

	ns.NullTime.Valid = true

	return nil
}

type Language int

const (
	English Language = iota + 1
	Spanish
)

func (l Language) String() string {
	return [...]string{
		English: "english",
		Spanish: "spanish",
	}[l]
}

// Scan implements the Scanner interface.
func (l *Language) Scan(value interface{}) error {
	switch value.(type) {
	case nil:
		*l = English
		return nil
	case string:
		v := value.(string)
		if v == "english" {
			*l = English
		} else if v == "spanish" {
			*l = Spanish
		}
		return nil
	case []byte:
		v := string(value.([]byte))
		if v == "english" {
			*l = English
		} else if v == "spanish" {
			*l = Spanish
		}
		return nil
	case int:
		v := value.(int)
		if v == int(English) {
			*l = English
		} else if v == int(Spanish) {
			*l = Spanish
		}
		return nil
	default:
		return fmt.Errorf("unrecognized value %v (type: %[1]T)", value)
	}
}

// Value implements the driver Valuer interface.
func (l Language) Value() (driver.Value, error) {
	return l.String(), nil
}

// MarshalJSON marshals the NullInt64 as null or the nested int64.
func (l Language) MarshalJSON() ([]byte, error) {
	if int(l) == 0 {
		return json.Marshal(nil)
	}

	return json.Marshal(l.String())
}

// UnmarshalJSON unmarshals a NullInt64
func (l *Language) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte(`"english"`)) {
		*l = English
		return nil
	} else if bytes.Equal(b, []byte(`"spanish"`)) {
		*l = Spanish
		return nil
	} else {
		var i int
		if err := json.Unmarshal(b, &i); err != nil {
			return err
		}
		*l = Language(i)
		return nil
	}
}
