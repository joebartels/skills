package accountlookup_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"example.com/accountlookup"
)

type backendFunc func(string) (string, error)

func (f backendFunc) Find(key string) (string, error) { return f(key) }

func TestLookupRejectsBlankKey(t *testing.T) {
	for _, key := range []string{"", " ", "\t\r\n", "\u00a0\u2003"} {
		t.Run(fmt.Sprintf("%q", key), func(t *testing.T) {
			called := false
			s := accountlookup.Service{Backend: backendFunc(func(string) (string, error) {
				called = true
				return "account", nil
			})}
			value, err := s.Lookup(key)
			if !errors.Is(err, accountlookup.ErrInvalidKey) {
				t.Errorf("error = %v, want ErrInvalidKey classification", err)
			}
			if value != "" {
				t.Errorf("value = %q, want empty on failure", value)
			}
			if called {
				t.Error("blank key reached backend")
			}
		})
	}
}

func TestLookupPreservesNonblankKey(t *testing.T) {
	for _, key := range []string{"u", " u ", "\u00a0u\u2003"} {
		t.Run(fmt.Sprintf("%q", key), func(t *testing.T) {
			calls := 0
			s := accountlookup.Service{Backend: backendFunc(func(got string) (string, error) {
				calls++
				if got != key {
					t.Errorf("backend key = %q, want %q", got, key)
				}
				return "account", nil
			})}
			value, err := s.Lookup(key)
			if value != "account" || err != nil {
				t.Errorf("Lookup = (%q, %v), want (account, nil)", value, err)
			}
			if calls != 1 {
				t.Errorf("backend calls = %d, want 1", calls)
			}
		})
	}
}

func TestLookupMissingIdentity(t *testing.T) {
	for _, backendErr := range []error{
		accountlookup.BackendMissing,
		fmt.Errorf("private SQL: %w", accountlookup.BackendMissing),
		errors.Join(errors.New("private password"), accountlookup.BackendMissing),
	} {
		s := accountlookup.Service{Backend: backendFunc(func(string) (string, error) {
			return "partial account", backendErr
		})}
		value, err := s.Lookup("u")
		if err != accountlookup.ErrMissing {
			t.Errorf("error = %v, want ErrMissing identity", err)
		}
		if value != "" {
			t.Errorf("value = %q, want empty on failure", value)
		}
	}
}

type backendFailure struct{}

func (*backendFailure) Error() string { return "password=secret; SELECT account" }

type noncomparableFailure []string

func (noncomparableFailure) Error() string { return "password=secret; SELECT account" }

func TestLookupUnavailableContextAndIsolation(t *testing.T) {
	for _, backendErr := range []error{
		&backendFailure{},
		noncomparableFailure{"private"},
		fmt.Errorf("backend wrapper: %w", &backendFailure{}),
	} {
		s := accountlookup.Service{Backend: backendFunc(func(string) (string, error) {
			return "partial account", backendErr
		})}
		key := "u\n\"admin"
		value, err := s.Lookup(key)
		if !errors.Is(err, accountlookup.ErrUnavailable) {
			t.Fatalf("error = %v, want ErrUnavailable classification", err)
		}
		if value != "" {
			t.Errorf("value = %q, want empty on failure", value)
		}
		if !strings.Contains(err.Error(), "lookup") || !strings.Contains(err.Error(), `"u\n\"admin"`) {
			t.Errorf("error = %q, want operation and quoted key context", err)
		}
		if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "SELECT") || strings.ContainsAny(err.Error(), "\r\n") {
			t.Errorf("error exposes sensitive diagnostics or raw line breaks: %q", err)
		}
		if errors.Is(err, backendErr) {
			t.Errorf("error exposes backend identity: %v", err)
		}
		var typed *backendFailure
		if errors.As(err, &typed) {
			t.Errorf("error exposes backend type: %T", typed)
		}
		var noncomparable noncomparableFailure
		if errors.As(err, &noncomparable) {
			t.Errorf("error exposes backend type: %T", noncomparable)
		}
	}
}
