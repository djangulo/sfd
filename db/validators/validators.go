//Package validators contains functions to validate input data.
// Most, if not all functions mirror the db.Driver interface in name and
// parameters.
package validators

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/djangulo/sfd/db/models"
	"github.com/gofrs/uuid"
)

const (
	// ByteLength 255
	ByteLength = 255
	dateFormat = "2006-01-02 15:04:05"
)

var (
	ErrCannotBeEmpty       = errors.New("cannot be empty")
	ErrNilPointer          = errors.New("pointer is nil")
	ErrAtLeast1Number      = errors.New("must contait at least one (1) number")
	ErrAtLeast1Uppercase   = errors.New("must contain at least one (1) uppercase letter")
	ErrAtLeast1Lowercase   = errors.New("must contain at least 1 lowercase letter")
	ErrAtLeast1SpecialChar = errors.New(`must contain at least 1 symbol (!@#$%^&*(){}[]/=?+-_,.<>;:'\",)`)
	ErrPhoneReNoMatch      = errors.New(`only digits [1-9]; separators [().-]; spaces; and the words "ext" or "x" for extension`)
)

func ErrCannotBeLessThan(n int) error {
	return fmt.Errorf("cannot be less than %d", n)
}

func ErrCannotBeGreaterThan(n int) error {
	return fmt.Errorf("cannot be greater than %d", n)
}

func ErrCannotBeShorterThan(n int) error {
	return fmt.Errorf("cannot be shorter than %d characters", n)
}

func ErrCannotBeLongerThan(n int) error {
	return fmt.Errorf("cannot be longer than %d characters", n)
}

func ErrCannotBeAfter(bound time.Time) error {
	return fmt.Errorf("cannot be after %s", bound.Format(dateFormat))
}

func ErrCannotBeBefore(bound time.Time) error {
	return fmt.Errorf("cannot be before %s", bound.Format(dateFormat))
}

// Errors is a struct that implements the error interface.
type Errors struct {
	Values map[string][]error
}

func (e *Errors) Render(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	return nil
}

// NewErrors initializes and returns a *Errors.
func NewErrors() *Errors {
	return &Errors{Values: make(map[string][]error)}
}

// Add adds err under key.
func (e *Errors) Len() int {
	if e != nil {
		if e.Values != nil {
			return len(e.Values)
		}
	}
	return 0
}

// Add adds err under key.
func (e *Errors) Add(key string, errs ...error) {
	if v := e.Values[key]; v == nil {
		e.Values[key] = make([]error, 0)
	}
	for _, err := range errs {
		e.Values[key] = append(e.Values[key], err)
	}
}

// Has returns whether Errors has any errors under key.
func (f *Errors) Has(key string) bool {
	if v, ok := f.Values[key]; ok {
		return len(v) > 0
	}
	return false
}

// Error implements the error interface.
func (e *Errors) Error() string {
	if e == nil {
		return "<nil>"
	}
	var errs = make([]string, 0)
	for k, v := range e.Values {
		errs = append(errs, fmt.Sprintf("%s:%v", k, v))
	}
	return strings.Join(errs, ",")
}

// String implements the fmt.Stringer interface.
func (e *Errors) String() string {
	return e.Error()
}

// AsList returns a ul>li of all the errors under key.
func (e *Errors) AsList(key string) template.HTML {
	var b strings.Builder

	b.WriteString("<ul>\n")
	if v, ok := e.Values[key]; ok {
		for i, err := range v {
			b.WriteString(
				fmt.Sprintf(
					`<li id="%s-error-%d">%s</li>`,
					key,
					i+1,
					err.Error(),
				),
			)
		}
	}
	b.WriteString("</ul>")

	return template.HTML(b.String())
}

func (e Errors) MarshalJSON() ([]byte, error) {
	var result = make(map[string]interface{})
	for k, v := range e.Values {
		s := make([]string, 0)
		for _, err := range v {
			s = append(s, err.Error())
		}
		result[k] = s
	}
	return json.Marshal(result)
}

// Min assert value is greater than n.
func Min(value, n int) error {
	if value < n {
		return ErrCannotBeLessThan(n)
	}
	return nil
}

// Max assert value is less than n.
func Max(value, n int) error {
	if value >= n {
		return ErrCannotBeGreaterThan(n)
	}
	return nil
}

// MinLength assert length of string value is greater than n.
func MinLength(value string, n int) error {
	if len(value) < n {
		return ErrCannotBeShorterThan(n)
	}
	return nil
}

// MaxLength assert length of string value is less than n.
func MaxLength(value string, n int) error {
	if len(value) >= n {
		return ErrCannotBeLongerThan(n)
	}
	return nil
}

// NotEmpty assert string value is not empty.
func NotEmpty(value string) error {
	if value == "" {
		return ErrCannotBeEmpty
	}
	return nil
}

// NotNil assert ptr is not a nil pointer.
func NotNil(ptr interface{}) error {
	v := reflect.ValueOf(ptr)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if v.IsNil() {
			return ErrNilPointer
		}
	default:
		return nil
	}
	return nil
}

var (
	reMu     sync.Mutex
	storedRe map[string]*regexp.Regexp
)

// getRe retrieves the cached *regexp.Regexp.
func getRE(re string) (*regexp.Regexp, error) {
	reMu.Lock()
	defer reMu.Unlock()

	if r, ok := storedRe[re]; ok {
		return r, nil
	}
	r, err := regexp.Compile(re)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Regex assert the string value matches re.
func Regex(value, re string) error {
	r, err := getRE(re)
	if err != nil {
		return err
	}
	if !r.MatchString(value) {
		return fmt.Errorf("%q does not match %q", value, re)
	}
	return nil
}

// Before asserts time t is before bound.
func Before(t, bound time.Time) error {
	if t.After(bound) {
		return ErrCannotBeAfter(bound)
	}
	return nil
}

// After asserts time t is after bound.
func After(t, bound time.Time) error {
	if t.Before(bound) {
		return ErrCannotBeBefore(bound)
	}
	return nil
}

// Email validates email s os not empty and less than 255.
func Email(s string) []error {
	e := make([]error, 0)

	if err := NotEmpty(s); err != nil {
		e = append(e, err)
	}
	if err := MaxLength(s, ByteLength); err != nil {
		e = append(e, err)
	}
	if len(e) > 0 {
		return e
	}
	return nil
}

// Password asserts the password s meets the following criteria:
//   - At least 1 lowercase letter.
//   - At least 1 uppercase letter.
//   - At least 1 number .
//   - At least 1 symbol ((!@#$%^&*(){}[]/=?+-_,.<>;:'\",)).
//   - At least 8 characters long.
func Password(s string) []error {
	e := make([]error, 0)

	num, lower, upper, symbol := false, false, false, false
	for _, r := range s {
		switch {
		case unicode.IsNumber(r):
			num = true
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			symbol = true
		}
	}
	if err := MinLength(s, 8); err != nil {
		e = append(e, err)
	}
	if !num {
		e = append(e, ErrAtLeast1Number)
	}
	if !upper {
		e = append(e, ErrAtLeast1Uppercase)
	}
	if !lower {
		e = append(e, ErrAtLeast1Lowercase)
	}
	if !symbol {
		e = append(e, ErrAtLeast1SpecialChar)
	}
	if len(e) > 0 {
		return e
	}
	return nil
}

func UUID(id *uuid.UUID) []error {
	e := make([]error, 0)

	if err := NotNil(id); err != nil {
		e = append(e, err)
	}
	if *id == uuid.Nil {
		e = append(e, fmt.Errorf("%v: uuid is nil", models.ErrInvalidInput.Error()))
	}
	if len(e) > 0 {
		return e
	}
	return nil
}

const (
	phoneReStr = `[\d\-\.\ \(\)+]+(x|ext)*\d+`
)

func PhoneNumber(p *models.PhoneNumber) []error {
	e := make([]error, 0)

	if err := NotNil(p); err != nil {
		e = append(e, err)
	}
	if err := NotEmpty(p.Number); err != nil {
		e = append(e, fmt.Errorf("numero: %v", err))
	}
	if err := UUID(&p.ID); err != nil {
		for _, err := range err {
			e = append(e, err)
		}
	}
	if err := Regex(p.Number, phoneReStr); err != nil {
		e = append(e, ErrPhoneReNoMatch)
	}
	if len(e) > 0 {
		return e
	}
	return nil
}

var cleanPhoneRegex = regexp.MustCompile(`[^\d\+x]`)

func cleanPhone(number string) string {
	number = cleanPhoneRegex.ReplaceAllString(number, "")
	return strings.TrimSpace(strings.ToLower(number))
}

// func ProfilePicture(p *models.ProfilePicture) []error {
// 	e := make([]error, 0)

// 	if err := ID(&p.ID); err != nil {
// 		for _, err := range err {
// 			e = append(e, err)
// 		}
// 	}
// 	if err := NotEmpty(p.Path); err != nil {
// 		e = append(e, err)
// 	}
// 	if err := NotEmpty(p.OriginalFilename); err != nil {
// 		e = append(e, err)
// 	}
// 	if err := NotEmpty(p.FileExt); err != nil {
// 		e = append(e, err)
// 	}

// 	if len(e) > 0 {
// 		return e
// 	}
// 	return nil
// }

// func Item(item *models.Item) []error {
// 	e := make([]error, 0)
// 	if err := ID(&item.ID); err != nil {
// 		for _, err := range err {
// 			e = append(e, err)
// 		}
// 	}
// 	if err := NotEmpty(item.Name); err != nil {
// 		e = append(e, err)
// 	}
// 	if err := MaxLength(item.Name, ByteLength); err != nil {
// 		e = append(e, err)
// 	}
// 	if err := NotEmpty(item.Slug); err != nil {
// 		e = append(e, err)
// 	}
// 	if err := MaxLength(item.Slug, ByteLength); err != nil {
// 		e = append(e, err)
// 	}
// 	if err := Min(int(item.MinBid.AsInt()), 0); err != nil {
// 		e = append(e, fmt.Errorf("apuesta mínima: no puede ser menor que 0"))
// 	}
// 	if item.MinBid.Gte(item.MaxBid) {
// 		e = append(e, fmt.Errorf("min / max: apuesta mínima no puede ser mayor o igual que apuesta máxima"))
// 	}

// 	if len(e) > 0 {
// 		return e
// 	}
// 	return nil
// }
