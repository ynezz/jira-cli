package md

import (
	"regexp"

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

// typedPanelPattern matches a complete typed panel block: {type}...{type} or {type:attrs}...{type}.
// Uses (?s) for DOTALL mode so . matches newlines.
// Capture groups: 1=panel type, 2=optional attributes, 3=content, 4=closing type.
var typedPanelPattern = regexp.MustCompile(`(?s)\{(info|warning|note|tip|error|success)(?::([^}]*))?\}(.*?)\{(info|warning|note|tip|error|success)\}`)

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

// ToJiraMD translates CommonMark to Jira flavored markdown.
func ToJiraMD(md string) string {
	if md == "" {
		return md
	}

	// Convert typed panels ({info}, {warning}, etc.) to {panel:bgColor=...} format
	md = convertTypedPanels(md)

	renderer := &cf.Renderer{Flags: cf.IgnoreMacroEscaping}
	r := bf.New(bf.WithRenderer(renderer), bf.WithExtensions(bf.CommonExtensions))

	return string(renderer.Render(r.Parse([]byte(md))))
}

// FromJiraMD translates Jira flavored markdown to CommonMark.
func FromJiraMD(jfm string) string {
	return jirawiki.Parse(jfm)
}
