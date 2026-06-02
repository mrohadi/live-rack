package domain

import "fmt"

// RestockTaskTitle is the canonical title for an auto-generated restock task.
// Stable by SKU so duplicate open tasks can be deduplicated by title match.
func RestockTaskTitle(sku string) string {
	return fmt.Sprintf("Restock %s", sku)
}
