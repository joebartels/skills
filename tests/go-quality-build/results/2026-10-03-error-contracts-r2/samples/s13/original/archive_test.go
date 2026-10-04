package entryzip

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestWriteArchive(t *testing.T) {
	var output bytes.Buffer
	if err := WriteArchive(&output, []Entry{{Name: "a.txt", Body: strings.NewReader("alpha")}}); err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if len(reader.File) != 1 || reader.File[0].Name != "a.txt" {
		t.Fatalf("entries = %v", reader.File)
	}
	body, err := reader.File[0].Open()
	if err != nil {
		t.Fatal(err)
	}
	defer body.Close()
	got, err := io.ReadAll(body)
	if err != nil || string(got) != "alpha" {
		t.Fatalf("body = %q, error = %v", got, err)
	}
}
