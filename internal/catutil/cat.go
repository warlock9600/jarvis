package catutil

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

type FileView struct {
	Path    string `json:"path"`
	Lexer   string `json:"lexer"`
	Binary  bool   `json:"binary"`
	Content string `json:"content"`
}

func ReadFile(path string) (FileView, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return FileView{}, err
	}
	binary := isProbablyBinary(b)
	content := string(b)
	lexer := detectLexer(path, content)
	return FileView{Path: path, Lexer: lexer, Binary: binary, Content: content}, nil
}

func RenderHighlighted(view FileView, styleName string, lineNumbers bool) (string, error) {
	if isMarkdownView(view) {
		return renderMarkdown(view.Content, styleName), nil
	}

	if styleName == "" {
		styleName = "monokai"
	}
	lexer := lexers.Get(view.Lexer)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		return "", fmt.Errorf("terminal formatter not available")
	}
	if lineNumbers {
		formatter = formatters.Get("terminal16m")
		if formatter == nil {
			formatter = formatters.Fallback
		}
	}
	iterator, err := lexer.Tokenise(nil, view.Content)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func isMarkdownView(view FileView) bool {
	ext := strings.ToLower(filepath.Ext(view.Path))
	if ext == ".md" || ext == ".markdown" {
		return true
	}
	lx := strings.ToLower(strings.TrimSpace(view.Lexer))
	return strings.Contains(lx, "markdown") || lx == "md"
}

type mdPalette struct {
	Header string
	List   string
	Quote  string
	Fence  string
	Code   string
	Link   string
	Reset  string
}

func paletteByStyle(styleName string) mdPalette {
	s := strings.ToLower(strings.TrimSpace(styleName))
	if s == "" {
		s = "monokai"
	}
	switch s {
	case "dracula":
		return mdPalette{
			Header: "\033[1;95m",
			List:   "\033[96m",
			Quote:  "\033[2;37m",
			Fence:  "\033[95m",
			Code:   "\033[92m",
			Link:   "\033[94m",
			Reset:  "\033[0m",
		}
	case "nord":
		return mdPalette{
			Header: "\033[1;96m",
			List:   "\033[94m",
			Quote:  "\033[2;37m",
			Fence:  "\033[94m",
			Code:   "\033[92m",
			Link:   "\033[96m",
			Reset:  "\033[0m",
		}
	case "github", "solarized-light", "vs":
		return mdPalette{
			Header: "\033[1;34m",
			List:   "\033[35m",
			Quote:  "\033[2;90m",
			Fence:  "\033[34m",
			Code:   "\033[32m",
			Link:   "\033[36m",
			Reset:  "\033[0m",
		}
	default: // monokai, solarized-dark, onedark, etc.
		return mdPalette{
			Header: "\033[1;96m",
			List:   "\033[93m",
			Quote:  "\033[2;37m",
			Fence:  "\033[95m",
			Code:   "\033[92m",
			Link:   "\033[94m",
			Reset:  "\033[0m",
		}
	}
}

func renderMarkdown(content string, styleName string) string {
	p := paletteByStyle(styleName)
	var b strings.Builder
	lines := strings.Split(content, "\n")
	inFence := false
	fenceRe := regexp.MustCompile("^\\s*```")
	headerRe := regexp.MustCompile(`^\s{0,3}#{1,6}\s+`)
	listRe := regexp.MustCompile(`^\s*([-*+]\s+|\d+\.\s+)`)
	quoteRe := regexp.MustCompile(`^\s*>+\s?`)
	inlineCodeRe := regexp.MustCompile("`[^`]+`")
	linkRe := regexp.MustCompile(`\[[^\]]+\]\([^)]+\)`)

	for i, line := range lines {
		rendered := line
		switch {
		case fenceRe.MatchString(line):
			inFence = !inFence
			rendered = p.Fence + line + p.Reset
		case inFence:
			rendered = p.Code + line + p.Reset
		case headerRe.MatchString(line):
			rendered = p.Header + line + p.Reset
		case quoteRe.MatchString(line):
			rendered = p.Quote + line + p.Reset
		case listRe.MatchString(line):
			rendered = p.List + line + p.Reset
		default:
			rendered = inlineCodeRe.ReplaceAllStringFunc(rendered, func(x string) string {
				return p.Code + x + p.Reset
			})
			rendered = linkRe.ReplaceAllStringFunc(rendered, func(x string) string {
				return p.Link + x + p.Reset
			})
		}
		b.WriteString(rendered)
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func detectLexer(path, content string) string {
	base := filepath.Base(path)
	if l := lexers.Match(base); l != nil {
		return l.Config().Name
	}
	if l := lexers.Analyse(content); l != nil {
		return l.Config().Name
	}
	return "plaintext"
}

func isProbablyBinary(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	if bytes.IndexByte(b, 0x00) >= 0 {
		return true
	}
	sample := b
	if len(sample) > 4096 {
		sample = sample[:4096]
	}
	nonPrintable := 0
	for _, c := range sample {
		if c == '\n' || c == '\r' || c == '\t' {
			continue
		}
		if c < 0x20 || c == 0x7f {
			nonPrintable++
		}
	}
	return float64(nonPrintable)/float64(len(sample)) > 0.20
}

func JoinWithHeader(outputs []string) string {
	if len(outputs) == 0 {
		return ""
	}
	if len(outputs) == 1 {
		return outputs[0]
	}
	return strings.Join(outputs, "\n")
}
