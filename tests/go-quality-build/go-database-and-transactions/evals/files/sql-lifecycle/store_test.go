package reportworker

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
)

func TestMaterialize(t *testing.T) {
	db := sql.OpenDB(happyConnector{})
	defer db.Close()
	got, err := New(db).Materialize(context.Background(), "acct-7", 3)
	want := Report{Account: "acct-7", Currency: "USD", Entries: 2, Total: 12}
	if err != nil || got != want {
		t.Fatalf("Materialize = %#v, %v; want %#v", got, err, want)
	}
}

func TestInvalidLimit(t *testing.T) {
	if _, err := New(nil).Materialize(context.Background(), "acct-7", 0); err == nil {
		t.Fatal("zero sample limit accepted")
	}
}

type happyConnector struct{}

func (happyConnector) Connect(context.Context) (driver.Conn, error) { return happyConn{}, nil }
func (happyConnector) Driver() driver.Driver                        { return happyDriver{} }

type happyDriver struct{}

func (happyDriver) Open(string) (driver.Conn, error) { return happyConn{}, nil }

type happyConn struct{}

func (happyConn) Prepare(string) (driver.Stmt, error)                          { return nil, errors.New("prepare not used") }
func (happyConn) Close() error                                                 { return nil }
func (happyConn) Begin() (driver.Tx, error)                                    { return happyTx{}, nil }
func (happyConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) { return happyTx{}, nil }
func (happyConn) QueryContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
	if q == currencySQL {
		return &happyRows{columns: []string{"currency"}, values: [][]driver.Value{{"USD"}}}, nil
	}
	return &happyRows{columns: []string{"amount"}, values: [][]driver.Value{{int64(5)}, {int64(7)}}}, nil
}
func (happyConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

type happyTx struct{}

func (happyTx) Commit() error   { return nil }
func (happyTx) Rollback() error { return nil }

type happyRows struct {
	columns []string
	values  [][]driver.Value
}

func (r *happyRows) Columns() []string { return r.columns }
func (r *happyRows) Close() error      { return nil }
func (r *happyRows) Next(dest []driver.Value) error {
	if len(r.values) == 0 {
		return io.EOF
	}
	copy(dest, r.values[0])
	r.values = r.values[1:]
	return nil
}
