package measure

import (
	"errors"
	"testing"
)

func TestFormatSeparation(t *testing.T) {
	strict := Validate
	legacy := ValidateLegacy
	m := Measurement{Value: 4, Legacy: true}
	if !errors.Is(strict(m), errInvalid) || legacy(m) != nil {
		t.Fatal("only the legacy path should accept old records")
	}
	m.Legacy = false
	if strict(m) != nil || legacy(m) != nil {
		t.Fatal("both paths should accept current records")
	}
}

func TestWithValue(t *testing.T) {
	m := Measurement{Value: 2}
	if m.WithValue(3).Value != 3 || m.Value != 2 {
		t.Fatal("WithValue must change only the copy")
	}
}

func TestTotal(t *testing.T) {
	got, err := Total([]Measurement{{Value: 2}, {Value: 3, Legacy: true}})
	if err != nil || got.Value != 5 {
		t.Fatalf("Total=%+v, %v", got, err)
	}
	strict := Validate
	if err := strict(got); err != nil {
		t.Fatalf("Total must produce current format: %v", err)
	}
}
