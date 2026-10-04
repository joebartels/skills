package accountlookup_test

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"example.com/accountlookup"
)

type backendFunc func(string) (string, error)

func (f backendFunc) Find(key string) (string, error) { return f(key) }

type diagnosticError struct{ text string }

func (e *diagnosticError) Error() string { return e.text }

type sliceError []string

func (e sliceError) Error() string { return strings.Join(e, " ") }

func TestLookupRejectsBlankKey(t *testing.T) {
	for _, key := range []string{"", " \t\n\r", "\u00a0\u2003"} {
		t.Run(strconv.Quote(key), func(t *testing.T) {
			calls := 0
			service := accountlookup.Service{Backend: backendFunc(func(string) (string, error) {
				calls++
				return "unexpected account", nil
			})}
			value, err := service.Lookup(key)
			if !errors.Is(err, accountlookup.ErrInvalidKey) {
				t.Errorf("Lookup error = %v, want ErrInvalidKey classification", err)
			}
			if value != "" {
				t.Errorf("Lookup value = %q, want empty on failure", value)
			}
			if calls != 0 {
				t.Errorf("backend calls = %d, want 0", calls)
			}
		})
	}
}

func TestMissing(t *testing.T) {
	for _, backendErr := range []error{
		accountlookup.BackendMissing,
		fmt.Errorf("backend query: %w", accountlookup.BackendMissing),
		errors.Join(accountlookup.BackendMissing, errors.New("password=secret")),
	} {
		service := accountlookup.Service{Backend: backendFunc(func(string) (string, error) {
			return "unusable account", backendErr
		})}
		value, err := service.Lookup("u")
		if err != accountlookup.ErrMissing {
			t.Errorf("Lookup error = %v, want direct ErrMissing", err)
		}
		if value != "" {
			t.Errorf("Lookup value = %q, want empty on failure", value)
		}
	}
}

func TestLookupUnavailable(t *testing.T) {
	const secret = "password=secret SELECT * FROM accounts"
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"direct", &diagnosticError{text: secret}},
		{"wrapped", fmt.Errorf("query: %w", &diagnosticError{text: secret})},
		{"noncomparable", sliceError{secret}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const key = "account\"\n\r\t42"
			service := accountlookup.Service{Backend: backendFunc(func(gotKey string) (string, error) {
				if gotKey != key {
					t.Errorf("backend key = %q, want %q", gotKey, key)
				}
				return "unusable account", tc.err
			})}
			value, err := service.Lookup(key)
			if !errors.Is(err, accountlookup.ErrUnavailable) {
				t.Fatalf("Lookup error = %v, want ErrUnavailable classification", err)
			}
			if value != "" {
				t.Errorf("Lookup value = %q, want empty on failure", value)
			}
			if !strings.Contains(err.Error(), "lookup") || !strings.Contains(err.Error(), strconv.Quote(key)) {
				t.Errorf("Lookup error = %q, want lookup and quoted key context", err)
			}
			if strings.Contains(err.Error(), secret) || strings.ContainsAny(err.Error(), "\n\r\t") {
				t.Errorf("Lookup error exposes sensitive text or unescaped controls: %q", err)
			}
			if errors.Is(err, tc.err) {
				t.Error("Lookup error exposes backend identity")
			}
			var diagnostic *diagnosticError
			if errors.As(err, &diagnostic) {
				t.Error("Lookup error exposes backend diagnostic type")
			}
			var noncomparable sliceError
			if errors.As(err, &noncomparable) {
				t.Error("Lookup error exposes backend slice error type")
			}
		})
	}
}

func TestLookupSuccessPreservesKey(t *testing.T) {
	const key = " \tu\n "
	calls := 0
	service := accountlookup.Service{Backend: backendFunc(func(gotKey string) (string, error) {
		calls++
		if gotKey != key {
			t.Errorf("backend key = %q, want unchanged %q", gotKey, key)
		}
		return "account", nil
	})}
	value, err := service.Lookup(key)
	if value != "account" || err != nil {
		t.Errorf("Lookup = (%q, %v), want (account, nil)", value, err)
	}
	if calls != 1 {
		t.Errorf("backend calls = %d, want 1", calls)
	}
}
