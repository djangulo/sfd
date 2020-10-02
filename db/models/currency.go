package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
)

// Currency type to represent money. Cents handled as integer/100; e.g.
// 100.00 == 10000
type Currency int64

// Add returns a new Currency with the added value of c and v. Integer conversion
// rules apply for v (1000 == 10.00, 100 = 1.00, 10 = 0.10, 1 = 0.11).
func (c Currency) Add(v interface{}) Currency {
	switch v.(type) {
	case Currency:
		return c + v.(Currency)
	case *Currency:
		return c + *(v.(*Currency))
	case int:
		return c + Currency(v.(int))
	case float64:
		return c + Currency(v.(float64))
	default:
		nc := MustCurrency(v)
		return c + nc
	}
}

// Sub returns a new Currency with the value of v substracted from c. Integer conversion
// rules apply for v (1000 == 10.00, 100 = 1.00, 10 = 0.10, 1 = 0.11).
func (c Currency) Sub(v interface{}) Currency {
	switch v.(type) {
	case Currency:
		return c - v.(Currency)
	case *Currency:
		return c - *(v.(*Currency))
	case int:
		return c - Currency(v.(int))
	case float64:
		return c - Currency(v.(float64))
	default:
		nc := MustCurrency(v)
		return c - nc
	}
}

// AsInt returns an int64 representation of the Currency.
func (c Currency) AsInt() int64 {
	return int64(c)
}

// AsFloat returns a float64 representation of the Currency.
func (c Currency) AsFloat() (f float64) {
	s := strconv.FormatInt(int64(c), 10)
	switch len(s) {
	case 0, 1:
		f, _ = strconv.ParseFloat("0.0"+s, 64)
	case 2:
		f, _ = strconv.ParseFloat("0."+s, 64)
	default:
		f, _ = strconv.ParseFloat(s[:len(s)-2]+"."+s[len(s)-2:], 64)
	}
	return f
}

// Gt greater than.
func (c Currency) Gt(v interface{}) bool {
	switch v.(type) {
	case Currency:
		return c > v.(Currency)
	case *Currency:
		return c > *(v.(*Currency))
	case int:
		return c > Currency(v.(int))
	case float64:
		return c > Currency(v.(float64))
	default:
		nc := MustCurrency(v)
		return c > nc
	}
}

// Gte greater than or equal.
func (c Currency) Gte(v interface{}) bool {
	switch v.(type) {
	case Currency:
		return c >= v.(Currency)
	case *Currency:
		return c >= *(v.(*Currency))
	case int:
		return c >= Currency(v.(int))
	case float64:
		return c >= Currency(v.(float64))
	default:
		nc := MustCurrency(v)
		return c >= nc
	}
}

// Lt greater than.
func (c Currency) Lt(v interface{}) bool {
	switch v.(type) {
	case Currency:
		return c < v.(Currency)
	case *Currency:
		return c < *(v.(*Currency))
	case int:
		return c < Currency(v.(int))
	case float64:
		return c < Currency(v.(float64))
	default:
		nc := MustCurrency(v)
		return c < nc
	}
}

// Lte greater than.
func (c Currency) Lte(v interface{}) bool {
	switch v.(type) {
	case Currency:
		return c <= v.(Currency)
	case *Currency:
		return c <= *(v.(*Currency))
	case int:
		return c <= Currency(v.(int))
	case float64:
		return c <= Currency(v.(float64))
	default:
		nc := MustCurrency(v)
		return c <= nc
	}
}

// Eq equal.
func (c Currency) Eq(v interface{}) bool {
	switch v.(type) {
	case Currency:
		return c == v.(Currency)
	case *Currency:
		return c == *(v.(*Currency))
	case int:
		return c == Currency(v.(int))
	case float64:
		return c == Currency(v.(float64))
	default:
		nc := MustCurrency(v)
		return c == nc
	}
}

// Neq not equal.
func (c Currency) Neq(v interface{}) bool {
	switch v.(type) {
	case Currency:
		return c != v.(Currency)
	case *Currency:
		return c != *(v.(*Currency))
	case int:
		return c != Currency(v.(int))
	case float64:
		return c != Currency(v.(float64))
	default:
		nc := MustCurrency(v)
		return c != nc
	}
}

// NewCurrency returns a new currency.
// - If v is a float, untyped or otherwise, it will be truncated to 2 decimal
//   digits, rounding down, and its total value multiplied by 100 before
//   conversion. e.g. NewCurrency(100.00) returns a currency of value 10000.
// - If v is an int of any size, it will be stored assuming the two least
//   significant zeroes are cents.
// - If it's a string, it will be parsed using strconv, any errors default
//   to Currency(0).
func NewCurrency(v interface{}) (Currency, error) {
	// c := &Currency{}
	switch v.(type) {
	case float64:
		s := strconv.FormatFloat(v.(float64), 'f', 2, 64)
		return NewCurrency(s)
	case float32:
		s := strconv.FormatFloat(float64(v.(float32)), 'f', 2, 32)
		return NewCurrency(s)
	case int:
		return Currency(v.(int)), nil
	case int8:
		return Currency(v.(int8)), nil
	case int16:
		return Currency(v.(int16)), nil
	case int32:
		return Currency(v.(int32)), nil
	case int64:
		return Currency(v.(int64)), nil
	case []uint8:
		s := string(v.([]uint8))
		return NewCurrency(s)
	case string:
		f, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return Currency(0), err
		}
		m := math.Round(f * 100)
		return Currency(int64(m)), nil
	default:
		return Currency(0), fmt.Errorf("unsupported type for currency: %T", v)
	}
}

// MustCurrency creates a new currency and panics on error.
func MustCurrency(v interface{}) Currency {
	c, err := NewCurrency(v)
	if err != nil {
		panic(err)
	}
	return c
}

func (c Currency) String() string {
	return fmt.Sprintf("%.2f", float64(c)/100.00)
}

// Value implements database/sql Valuer interface.
func (c Currency) Value() (driver.Value, error) {
	return driver.Value(float64(c) / 100), nil
}

// Scan implements the database/sql Scanner interface.
func (c *Currency) Scan(value interface{}) error {
	cur, err := NewCurrency(value)
	if err != nil {
		return err
	}
	*c = cur
	return nil
}

// MarshalJSON from the json.Marshaler interface. It's always stored as a float.
func (c Currency) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(c) / 100.00)
}

// UnmarshalJSON from the json.Unmarshaler interface
func (c *Currency) UnmarshalJSON(data []byte) (err error) {
	*c, err = NewCurrency(string(data))
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalForm implements the forms.Unmarshaler interface.
func (c *Currency) UnmarshalForm(form url.Values) error {
	v := form.Get("amount")
	if v == "" {
		return fmt.Errorf("\"amount\" not found")
	}
	cu, err := NewCurrency(v)
	if err != nil {
		return err
	}
	*c = cu
	return nil
}
