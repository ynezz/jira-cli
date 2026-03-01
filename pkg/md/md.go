package md

import (
	"regexp"
	"strings"
	"sync"

	cf "github.com/kentaro-m/blackfriday-confluence"
	bf "github.com/russross/blackfriday/v2"

	"github.com/ankitpokhrel/jira-cli/pkg/md/jirawiki"
)

// Panel type to background color mapping (from ADF spec).
var panelColors = map[string]string{
	"info":    "#deebff",
	"note":    "#eae6ff",
	"warning": "#fffae6",
	"tip":     "#e3fcef",
	"error":   "#ffebe6",
	"success": "#e3fcef",
}

var toJiraMDMu sync.Mutex

// typedPanelPattern matches a complete typed panel block: {type}...{type} or {type:attrs}...{type}.
// Uses (?s) for DOTALL mode so . matches newlines.
// Capture groups: 1=panel type, 2=optional attributes, 3=content, 4=closing type.
var typedPanelPattern = regexp.MustCompile(`(?s)\{(info|warning|note|tip|error|success)(?::([^}]*))?\}(.*?)\{(info|warning|note|tip|error|success)\}`)

// markdownListPattern matches a markdown list item with space indentation.
var markdownListPattern = regexp.MustCompile(`^([ ]*)([*+-]|\d+\.)\s+`)

// codeLanguagePattern matches renderer output line for typed code blocks.
var codeLanguagePattern = regexp.MustCompile(`^\{code:language=([^}\r\n]+)\}$`)

// typedCodePattern matches code macro opening tags that already use shorthand.
var typedCodePattern = regexp.MustCompile(`^\{code:[^}\r\n]+\}$`)

// minPanelSubmatches is the minimum number of submatches expected from typedPanelPattern,
// including the full match at index 0.
const minPanelSubmatches = 5

// convertTypedPanels converts typed wiki markup panels like {info}...{info}
// to {panel:bgColor=...}...{panel} format that Jira Cloud recognizes.
func convertTypedPanels(input string) string {
	return typedPanelPattern.ReplaceAllStringFunc(input, func(match string) string {
		submatch := typedPanelPattern.FindStringSubmatch(match)
		if len(submatch) < minPanelSubmatches {
			return match
		}

		panelType := submatch[1] // Opening tag type
		closingType := submatch[4]
		if panelType != closingType {
			return match
		}

		attrs := submatch[2]   // Optional attributes
		content := submatch[3] // Content between tags

		color, ok := panelColors[panelType]
		if !ok {
			return match
		}

		// Build the panel tag with bgColor
		openTag := "{panel:bgColor=" + color
		if attrs != "" {
			openTag += "|" + attrs
		}
		openTag += "}"

		return openTag + content + "{panel}"
	})
}

func normalizeCodeBlocks(input string) string {
	if input == "" {
		return input
	}

	hasTrailingNewline := strings.HasSuffix(input, "\n")
	trimmedInput := strings.TrimSuffix(input, "\n")
	lines := strings.Split(trimmedInput, "\n")

	inCodeBlock := false
	useNoFormat := false
	out := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if submatch := codeLanguagePattern.FindStringSubmatch(trimmed); len(submatch) == 2 {
			out = append(out, "{code:"+submatch[1]+"}")
			inCodeBlock = true
			useNoFormat = false
			continue
		}

		if typedCodePattern.MatchString(trimmed) {
			out = append(out, line)
			inCodeBlock = true
			useNoFormat = false
			continue
		}

		if trimmed == "{code}" {
			if inCodeBlock {
				if useNoFormat {
					out = append(out, "{noformat}")
				} else {
					out = append(out, "{code}")
				}
				inCodeBlock = false
				useNoFormat = false
			} else {
				out = append(out, "{noformat}")
				inCodeBlock = true
				useNoFormat = true
			}
			continue
		}

		out = append(out, line)
	}

	output := strings.Join(out, "\n")
	if hasTrailingNewline {
		output += "\n"
	}

	return output
}

func normalizeListIndentation(input string) string {
	if input == "" {
		return input
	}

	hasTrailingNewline := strings.HasSuffix(input, "\n")
	trimmedInput := strings.TrimSuffix(input, "\n")
	lines := strings.Split(trimmedInput, "\n")

	inFencedCode := false
	minIndent := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFencedCode = !inFencedCode
			continue
		}
		if inFencedCode {
			continue
		}

		match := markdownListPattern.FindStringSubmatch(line)
		if len(match) < 2 {
			continue
		}

		indent := len(match[1])
		if indent == 0 {
			continue
		}

		if minIndent == 0 || indent < minIndent {
			minIndent = indent
		}
	}

	// Preserve existing behavior for 4-space style lists and normalize only
	// the common 2-space nested list style.
	if minIndent != 2 {
		return input
	}

	inFencedCode = false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFencedCode = !inFencedCode
			continue
		}
		if inFencedCode {
			continue
		}

		match := markdownListPattern.FindStringSubmatch(line)
		if len(match) < 2 {
			continue
		}

		indent := len(match[1])
		if indent == 0 || indent%2 != 0 {
			continue
		}

		lines[i] = strings.Repeat(" ", indent*2) + line[indent:]
	}

	output := strings.Join(lines, "\n")
	if hasTrailingNewline {
		output += "\n"
	}

	return output
}

// ToJiraMD translates CommonMark to Jira flavored markdown.
func ToJiraMD(md string) string {
	if md == "" {
		return md
	}

	// Convert typed panels ({info}, {warning}, etc.) to {panel:bgColor=...} format
	md = convertTypedPanels(md)
	md = normalizeListIndentation(md)

	renderer := &cf.Renderer{Flags: cf.IgnoreMacroEscaping}
	r := bf.New(bf.WithRenderer(renderer), bf.WithExtensions(bf.CommonExtensions))

	// blackfriday-confluence uses shared package-level list state internally;
	// serialize rendering to avoid cross-goroutine state corruption.
	toJiraMDMu.Lock()
	defer toJiraMDMu.Unlock()

	output := string(renderer.Render(r.Parse([]byte(md))))
	return normalizeCodeBlocks(output)
}

// FromJiraMD translates Jira flavored markdown to CommonMark.
func FromJiraMD(jfm string) string {
	return jirawiki.Parse(jfm)
}
