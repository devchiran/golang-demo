package postgres

import "fmt"

func tableColumn(table, column string) string {
	return fmt.Sprintf("%s.%s", table, column)
}

func tableColumns(table string, columns []string) []string {
	cs := make([]string, 0, len(columns))
	for _, c := range columns {
		cs = append(cs, tableColumn(table, c))
	}
	return cs
}
