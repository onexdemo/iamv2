// Package table renders CLI tables.
package table

import (
	"io"
	"time"

	"github.com/olekukonko/tablewriter"
)

// Print writes a table with the given headers and rows.
func Print(out io.Writer, headers []string, rows [][]string) {
	t := tablewriter.NewWriter(out)
	t.SetHeader(headers)
	t.SetBorder(false)
	t.SetColumnSeparator("")
	t.SetHeaderLine(false)
	t.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	t.AppendBulk(rows)
	t.Render()
}

// FormatTime renders a timestamp in a readable form.
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format(time.DateTime)
}
