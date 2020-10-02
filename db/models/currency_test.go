package models

import (
	"fmt"
	"testing"
)

func TestNewCurrency(t *testing.T) {
	for _, test := range []struct {
		v    interface{}
		want Currency
		str  string
	}{
		{10.12, Currency(1012), "10.12"},
		{10.23432414, Currency(1023), "10.23"},
		{6000000.1223413, Currency(600000012), "6000000.12"},
		{float32(10.12), Currency(1012), "10.12"},
		{float32(10.23432414), Currency(1023), "10.23"},
		{float32(60000.1223413), Currency(6000012), "60000.12"},
		{10.00, Currency(1000), "10.00"},
		{10.0, Currency(1000), "10.00"},
		{11.01, Currency(1101), "11.01"},
		{1000, Currency(1000), "10.00"},
		{100, Currency(100), "1.00"},
		{1101, Currency(1101), "11.01"},
		{10, Currency(10), "0.10"},
		{20, Currency(20), "0.20"},
		{30, Currency(30), "0.30"},
		{int8(127), Currency(127), "1.27"},
		{int8(100), Currency(100), "1.00"},
		{int8(10), Currency(10), "0.10"},
		{int8(20), Currency(20), "0.20"},
		{int8(30), Currency(30), "0.30"},
		{int16(255), Currency(255), "2.55"},
		{int16(100), Currency(100), "1.00"},
		{int16(201), Currency(201), "2.01"},
		{int16(10), Currency(10), "0.10"},
		{int16(20), Currency(20), "0.20"},
		{int16(30), Currency(30), "0.30"},
		{[]uint8("9.55"), Currency(955), "9.55"},
		{[]uint8("0.55"), Currency(55), "0.55"},
		{[]uint8("12323.32"), Currency(1232332), "12323.32"},
		{[]byte("9.55"), Currency(955), "9.55"},
		{[]byte("0.55"), Currency(55), "0.55"},
		{[]byte("12323.32"), Currency(1232332), "12323.32"},
		{"9.55", Currency(955), "9.55"},
		{"0.55", Currency(55), "0.55"},
		{"12323.32", Currency(1232332), "12323.32"},
	} {
		t.Run(fmt.Sprintf("NewCurrency(%v)", test.v), func(t *testing.T) {
			t.Run("value", func(t *testing.T) {
				got, _ := NewCurrency(test.v)
				if got != test.want {
					t.Errorf("expected %d got %v", test.want, got)
				}
			})
			t.Run("string representation", func(t *testing.T) {
				curr, _ := NewCurrency(test.v)
				got := curr.String()
				if got != test.str {
					t.Errorf("expected %q got %q", test.str, got)
				}
			})
		})
	}
}

func TestCurrencyUnmarshalJSON(t *testing.T) {
	for _, test := range []struct {
		in   string
		want Currency
		str  string
	}{
		{"10.12", Currency(1012), "10.12"},
		{"10.23432414", Currency(1023), "10.23"},
		{"6000000.1223413", Currency(600000012), "6000000.12"},
		{"10.12", Currency(1012), "10.12"},
		{"10.23432414", Currency(1023), "10.23"},
		{"60000.1223413", Currency(6000012), "60000.12"},
		{"10.00", Currency(1000), "10.00"},
		{"10.0", Currency(1000), "10.00"},
		{"11.01", Currency(1101), "11.01"},
		{"1000", Currency(100000), "10.00"},
		{"2000", Currency(200000), "20.00"},
		{"3000", Currency(300000), "30.00"},
		{"10", Currency(1000), "10.00"},
		{"20", Currency(2000), "20.00"},
		{"30", Currency(3000), "30.00"},
	} {
		t.Run(fmt.Sprintf("UnmarshalJSON(%q)", test.in), func(t *testing.T) {
			c := Currency(0)
			err := (&c).UnmarshalJSON([]byte(test.in))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if c != test.want {
				t.Errorf("expected %d got %v", test.want, c)
			}
		})

	}
}

func TestCurrencyMarshalMethodsJSON(t *testing.T) {
	for _, test := range []struct {
		in   Currency
		want string
	}{
		{Currency(1012), "10.12"},
		{Currency(1023), "10.23"},
		{Currency(600000012), "6000000.12"},
		{Currency(1012), "10.12"},
		{Currency(1023), "10.23"},
		{Currency(6000012), "60000.12"},
		{Currency(1001), "10.01"},
		{Currency(1010), "10.1"},
		{Currency(1000), "10"},
		{Currency(1000), "10"},
		{Currency(1101), "11.01"},
		{Currency(10), "0.1"},
		{Currency(20), "0.2"},
		{Currency(30), "0.3"},
	} {
		t.Run(fmt.Sprintf("MarshalJSON(Currency(%d))", int(test.in)), func(t *testing.T) {
			b, err := test.in.MarshalJSON()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if string(b) != test.want {
				t.Errorf("expected %v got %v", test.want, string(b))
			}
		})

	}
}
