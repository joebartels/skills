package catalogue_test

import (
	"strings"
	"testing"

	"example.com/catalogue"
)

func TestImportSingleItem(t *testing.T) {
	calls := 0
	err := catalogue.Import(strings.NewReader("id,name,stock\n42,Bolt,7\n"), func(items []catalogue.Item) error {
		calls++
		if len(items) != 1 || items[0].ID != 42 || items[0].Name != "Bolt" || items[0].Stock == nil || *items[0].Stock != 7 {
			t.Fatalf("replacement = %#v", items)
		}
		return nil
	})
	if err != nil || calls != 1 {
		t.Fatalf("Import error = %v, replacement calls = %d", err, calls)
	}
}
