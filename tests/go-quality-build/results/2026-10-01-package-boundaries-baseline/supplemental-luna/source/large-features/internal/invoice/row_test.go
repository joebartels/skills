package invoice

import (
	"bytes"
	"reflect"
	"testing"
)

func TestRowCodecRoundTrip(t *testing.T) {
	want := Row{ID: "inv-7", Cents: 305}
	encoded, err := Encode(want)
	if err != nil || string(encoded) != "inv-7,305\n" {
		t.Fatalf("Encode = %q, %v", encoded, err)
	}
	got, err := Decode(string(encoded))
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Decode = %#v, %v", got, err)
	}
}

func TestDecodeRejectsInvalidRows(t *testing.T) {
	for _, line := range []string{"", ",4", "id,0", "id,-1", "id,nope", "id,2,3", "bad\r,4", "bad\n,4"} {
		if _, err := Decode(line); err == nil {
			t.Errorf("Decode(%q) succeeded", line)
		}
	}
}

func TestReadStopsAtFirstError(t *testing.T) {
	var got []Row
	err := Read(bytes.NewBufferString("one,10\nbad,0\ntwo,20\n"), func(row Row) error {
		got = append(got, row)
		return nil
	})
	if err == nil || !reflect.DeepEqual(got, []Row{{ID: "one", Cents: 10}}) {
		t.Fatalf("Read rows = %#v, error = %v", got, err)
	}
}
