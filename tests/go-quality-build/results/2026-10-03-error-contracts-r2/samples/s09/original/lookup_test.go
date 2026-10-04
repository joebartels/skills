package accountlookup

import "testing"

type missingBackend struct{}

func (missingBackend) Find(string) (string, error) { return "", BackendMissing }
func TestMissing(t *testing.T) {
	_, err := (Service{Backend: missingBackend{}}).Lookup("u")
	if err != ErrMissing {
		t.Fatal(err)
	}
}
