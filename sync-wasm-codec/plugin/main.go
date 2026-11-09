//go:build wasip1

// Command plugin is the sync-wasm-codec example's codec, for a line format
// of its own: id|name|price, a product a line, as "3|plum|12.00". Decode
// makes a row of a line, with an int id and a decimal price; encode writes
// a row back as such a line. Build it from the repository's root, by Go or
// by TinyGo:
//
//	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o sync-wasm-codec/prices.wasm ./sync-wasm-codec/plugin
//	tinygo build -target=wasip1 -buildmode=c-shared -o sync-wasm-codec/prices.wasm ./sync-wasm-codec/plugin
package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/lesismal/fit/plugin/wasm/guest"
)

func init() {
	guest.OnDecode(decode)
	guest.OnEncode(encode)
}

// decode makes a record of each line of data: the file source hands it a
// line at a time. A line not in the format fails, and the source keeps it
// as a message record holding the line and the error.
func decode(data []byte) error {
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			return errors.New(strconv.Quote(line) + " is not id|name|price")
		}
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return errors.New("the id " + strconv.Quote(parts[0]) + " is not a number")
		}
		if _, err := strconv.ParseFloat(parts[2], 64); err != nil {
			return errors.New("the price " + strconv.Quote(parts[2]) + " is not a number")
		}
		key := guest.Row{{Name: "id", Value: guest.Int(id)}}
		row := guest.Row{
			{Name: "id", Value: guest.Int(id)},
			{Name: "name", Value: guest.String(parts[1])},
			// A decimal keeps the price exact.
			{Name: "price", Value: guest.Decimal(parts[2])},
		}
		guest.Emit(0, &guest.Record{Op: "insert", Key: key.Value(), After: row.Value()})
	}
	return nil
}

// encode writes each record's row as a line; a row without an id fails,
// and the file sink hands that record to on_error.
func encode(batch []guest.Record) ([]byte, error) {
	var b []byte
	for _, r := range batch {
		row, err := guest.ParseRow(r.After)
		if err != nil || row == nil {
			return nil, errors.New("a record without a row")
		}
		idv, _ := row.Get("id")
		id, ok := guest.AsInt(idv)
		if !ok {
			// A message record of a line decode refused, say.
			return nil, errors.New("a row without an id")
		}
		namev, _ := row.Get("name")
		name, _ := guest.AsString(namev)
		b = strconv.AppendInt(b, id, 10)
		b = append(append(append(b, '|'), name...), '|')
		b = append(append(b, decimal(row)...), '\n')
	}
	return b, nil
}

// decimal returns the row's price, a decimal's digits, or the text of one
// held as a string.
func decimal(row guest.Row) string {
	v, _ := row.Get("price")
	if s, ok := guest.AsDecimal(v); ok {
		return s
	}
	s, _ := guest.AsString(v)
	return s
}

func main() {}
