package warehouse_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	warehouse "example.com/warehouse"
	_ "github.com/mattn/go-sqlite3"
)

func TestReserveExistingSingleLineWorkflow(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "warehouse.db")+"?_foreign_keys=on&_busy_timeout=1000&_txlock=immediate")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO customers VALUES ('shop'); INSERT INTO stock VALUES ('tea', 8)"); err != nil {
		t.Fatal(err)
	}
	s := warehouse.New(db)
	got, err := s.Reserve(context.Background(), warehouse.Request{ID: "r1", CustomerID: "shop", Lines: []warehouse.Line{{SKU: "tea", Quantity: 3}}})
	if err != nil {
		t.Fatal(err)
	}
	want := warehouse.Receipt{ID: "r1", Units: 3, Remaining: map[string]int{"tea": 5}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("receipt = %#v, want %#v", got, want)
	}
	available, err := s.Available(context.Background(), "tea")
	if err != nil || available != 5 {
		t.Fatalf("Available = %d, %v; want 5, nil", available, err)
	}
}
