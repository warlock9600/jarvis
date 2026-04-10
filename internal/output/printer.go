package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/term"
)

type Printer struct {
	JSON    bool
	NoColor bool
	IsTTY   bool
}

func New(jsonOut bool, noColor bool) *Printer {
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))
	return &Printer{JSON: jsonOut, NoColor: noColor, IsTTY: isTTY}
}

func (p *Printer) PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func (p *Printer) PrintTable(headers []string, rows [][]string) {
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	h := make(table.Row, len(headers))
	for i, x := range headers {
		h[i] = p.colorizeCell(x, text.FgMagenta)
	}
	tw.AppendHeader(h)
	for _, r := range rows {
		row := make(table.Row, len(r))
		for i, x := range r {
			row[i] = p.colorizeCell(x, text.FgGreen)
		}
		tw.AppendRow(row)
	}
	tw.SetStyle(table.StyleLight)
	tw.Style().Color.Header = text.Colors{}
	tw.Style().Color.Row = text.Colors{}
	tw.Style().Color.Footer = text.Colors{}
	tw.Render()
}

func (p *Printer) colorizeCell(v string, c text.Color) string {
	if p.NoColor || !p.IsTTY {
		return v
	}
	return text.Colors{c}.Sprint(v)
}

func (p *Printer) PrintKV(m map[string]string) {
	for k, v := range m {
		fmt.Fprintf(os.Stdout, "%s: %s\n", k, v)
	}
}
