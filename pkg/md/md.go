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

// typedPanelTagPattern matches a typed panel tag: {info} or {info:title=...}.
// Capture groups: 1=panel type, 2=optional attributes.
var typedPanelTagPattern = regexp.MustCompile(`\{(info|warning|note|tip|error|success)(?::([^}]*))?\}`)

// markdownListPattern matches a markdown list item with space indentation.
var markdownListPattern = regexp.MustCompile(`^([ ]*)([*+-]|\d+\.)\s+`)

// codeLanguagePattern matches renderer output line for typed code blocks.
var codeLanguagePattern = regexp.MustCompile(`^\{code:language=([^}\r\n]+)\}$`)

// typedCodePattern matches code macro opening tags that already use shorthand.
var typedCodePattern = regexp.MustCompile(`^\{code:[^}\r\n]+\}$`)

// convertTypedPanels converts typed wiki markup panels like {info}...{info}
// to {panel:bgColor=...}...{panel} format that Jira Cloud recognizes.
func convertTypedPanels(input string) string {
	type typedPanelTag struct {
		panelType string
		attrs     string
		start     int
		end       int
	}

	findInnermostPair := func(value string) (typedPanelTag, typedPanelTag, bool) {
		stack := make([]typedPanelTag, 0)
		matches := typedPanelTagPattern.FindAllStringSubmatchIndex(value, -1)
		for _, m := range matches {
			if len(m) < 6 {
				continue
			}

			tag := typedPanelTag{
				panelType: value[m[2]:m[3]],
				start:     m[0],
				end:       m[1],
			}
			if m[4] >= 0 && m[5] >= 0 {
				tag.attrs = value[m[4]:m[5]]
			}

			// Tags with explicit attributes are always treated as opening tags.
			if tag.attrs != "" {
				stack = append(stack, tag)
				continue
			}

			if len(stack) > 0 && stack[len(stack)-1].panelType == tag.panelType {
				return stack[len(stack)-1], tag, true
			}

			stack = append(stack, tag)
		}

		return typedPanelTag{}, typedPanelTag{}, false
	}

	for {
		open, close, ok := findInnermostPair(input)
		if !ok {
			return input
		}

		color, colorOK := panelColors[open.panelType]
		if !colorOK {
			return input
		}

		openTag := "{panel:bgColor=" + color
		if open.attrs != "" {
			openTag += "|" + open.attrs
		}
		openTag += "}"

		input = input[:open.start] + openTag + input[open.end:close.start] + "{panel}" + input[close.end:]
	}
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
	codeBlockStartIdx := -1
	out := make([]string, 0, len(lines))

	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)

		if submatch := codeLanguagePattern.FindStringSubmatch(trimmed); len(submatch) == 2 {
			out = append(out, "{code:"+submatch[1]+"}")
			inCodeBlock = true
			useNoFormat = false
			codeBlockStartIdx = len(out) - 1
			continue
		}

		if typedCodePattern.MatchString(trimmed) {
			out = append(out, line)
			inCodeBlock = true
			useNoFormat = false
			codeBlockStartIdx = len(out) - 1
			continue
		}

		if trimmed == "{code}" {
			if inCodeBlock {
				if shouldCloseCodeBlock(lines, idx) {
					if useNoFormat {
						out = append(out, "{noformat}")
					} else {
						out = append(out, "{code}")
					}
					inCodeBlock = false
					useNoFormat = false
					codeBlockStartIdx = -1
				} else {
					// Jira treats a literal `{code}` line as a closing macro in
					// code-tagged blocks. Fallback to noformat for this block so
					// the literal line remains intact in cloud rendering.
					if !useNoFormat {
						if codeBlockStartIdx >= 0 {
							out[codeBlockStartIdx] = "{noformat}"
						}
						useNoFormat = true
					}
					out = append(out, line)
				}
			} else {
				out = append(out, "{noformat}")
				inCodeBlock = true
				useNoFormat = true
				codeBlockStartIdx = len(out) - 1
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

func shouldCloseCodeBlock(lines []string, index int) bool {
	if index >= len(lines)-1 {
		return true
	}

	return strings.TrimSpace(lines[index+1]) == ""
}

func normalizeListIndentation(input string) string {
	if input == "" {
		return input
	}

	hasTrailingNewline := strings.HasSuffix(input, "\n")
	trimmedInput := strings.TrimSuffix(input, "\n")
	lines := strings.Split(trimmedInput, "\n")

	inFencedCode := false
	currentListIndentUnit := 0

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
		if trimmed == "" {
			currentListIndentUnit = 0
			continue
		}

		match := markdownListPattern.FindStringSubmatch(line)
		if len(match) < 2 {
			currentListIndentUnit = 0
			continue
		}

		indent := len(match[1])
		if indent == 0 {
			continue
		}

		if currentListIndentUnit == 0 {
			switch {
			case indent%4 == 0:
				// Preserve already-canonical 4-space nested list indentation.
				currentListIndentUnit = -1
				continue
			case indent%2 == 0 && indent%3 != 0:
				currentListIndentUnit = 2
			case indent%3 == 0 && indent%2 != 0:
				currentListIndentUnit = 3
			case indent%2 == 0:
				currentListIndentUnit = 2
			case indent%3 == 0:
				currentListIndentUnit = 3
			default:
				continue
			}
		}
		if currentListIndentUnit < 0 {
			continue
		}

		if indent%currentListIndentUnit != 0 {
			continue
		}

		normalizedIndent := (indent / currentListIndentUnit) * 4
		lines[i] = strings.Repeat(" ", normalizedIndent) + line[indent:]
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
