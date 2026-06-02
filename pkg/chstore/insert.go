package chstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// EncodeJSONEachRow serialises rows into ClickHouse's JSONEachRow format: one
// compact JSON object per line. Pure.
func EncodeJSONEachRow(rows []map[string]any) ([]byte, error) {
	var b strings.Builder
	for i, row := range rows {
		line, err := json.Marshal(row)
		if err != nil {
			return nil, fmt.Errorf("chstore: encode row %d: %w", i, err)
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return []byte(b.String()), nil
}

// Insert writes rows into table using JSONEachRow. A no-op on empty input.
func (c *Client) Insert(ctx context.Context, table string, rows []map[string]any) error {
	if len(rows) == 0 {
		return nil
	}
	body, err := EncodeJSONEachRow(rows)
	if err != nil {
		return err
	}
	stmt := "INSERT INTO " + table + " FORMAT JSONEachRow\n" + string(body)
	if err := c.Exec(ctx, stmt); err != nil {
		return fmt.Errorf("chstore: insert into %s: %w", table, err)
	}
	return nil
}
