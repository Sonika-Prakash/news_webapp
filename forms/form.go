package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type errs map[string][]string

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Form ...
type Form struct {
	url.Values
	Errors errs
}

func (e errs) Add(field, message string) {
	e[field] = append(e[field], message)
}

func (e errs) First(field string) string {
	errSlice := e[field]
	if len(errSlice) == 0 {
		return ""
	}
	return errSlice[0]
}

// New ...
func New(data url.Values) *Form {
	return &Form{
		data,
		make(errs),
	}
}

// Email gets the email address from the form values
func (f *Form) Email(field string) *Form {
	val := f.Get(field)
	if val == "" || !emailRegex.MatchString(val) {
		f.Errors.Add(field, "The value provided is not a valid email address")
	}
	return f
}

// MinLength checks the required min length criteria for a field in form
func (f *Form) MinLength(field string, d int) *Form {
	val := f.Get(field)
	if utf8.RuneCountInString(val) < d {
		f.Errors.Add(field, fmt.Sprintf("This %s is too short (minimum is %d characters)", field, d))
	}
	return f
}

// MaxLength checks the required min length criteria for a field in form
func (f *Form) MaxLength(field string, d int) *Form {
	val := f.Get(field)
	if val == "" {
		return f
	}
	if utf8.RuneCountInString(val) > d {
		f.Errors.Add(field, fmt.Sprintf("This %s is too long (maximum is %d characters)", field, d))
	}
	return f
}

// Required checks if a form field is filled correctly or not
func (f *Form) Required(fields ...string) *Form {
	for _, field := range fields {
		val := f.Get(field)
		if strings.TrimSpace(val) == "" {
			f.Errors.Add(field, fmt.Sprintf("%s is required", field))
		}
	}
	return f
}

// Valid validates the form by checking if there are no errors
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// Fail adds an error in the form
func (f *Form) Fail(field, msg string) {
	f.Errors.Add(field, msg)
}

// URL validates the URL
func (f *Form) URL(field string) *Form {
	val := f.Get(field)
	u, err := url.Parse(val)
	if err != nil || u.Scheme == "" || u.Host == "" {
		f.Errors.Add(field, fmt.Sprintf("%s is not a valid URL", field))
	}
	return f
}

// GetInt returns the int value for a field
func (f *Form) GetInt(field string) int {
	val, err := strconv.Atoi(f.Get(field))
	if err != nil {
		return 0
	}
	return val

}
