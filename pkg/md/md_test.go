package md

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToJiraMD(t *testing.T) {
	jfm := `# H1
Some _Markdown_ text.

## H2
Foobar.

### H3
Fuga

> quote

- - - -

**strong text**
~~strikethrough text~~
[Example Domain](http://www.example.com/)
![](https://path.to/image.jpg)

* list1
* list2
* list3

Paragraph

1. number1
2. number2
3. number3

|a  |b  |c  |
|---|---|---|
|1  |2  |3  |
|4  |5  |6  |

{panel:title=My Title}
**Subtitle**

Some text with a title
{panel}

` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("hello world")
}` + "```"

	expected := `h1. H1
Some _Markdown_ text.

h2. H2
Foobar.

h3. H3
Fuga

{quote}
quote

{quote}


----
*strong text*
-strikethrough text-
[Example Domain|http://www.example.com/]
!https://path.to/image.jpg!

* list1
* list2
* list3

Paragraph

# number1
# number2
# number3

||a||b||c||
|1|2|3|
|4|5|6|

{panel:title=My Title}
*Subtitle*

Some text with a title
{panel}

` + "```go" + `
package main

import "fmt"

func main\(\) {
    fmt.Println\("hello world"\)
}` + "```\n\n"

	assert.Equal(t, expected, ToJiraMD(jfm))
}

func TestConvertTypedPanels(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "info panel single line",
			input:    "{info}This is info content.{info}",
			expected: "{panel:bgColor=#deebff}This is info content.{panel}",
		},
		{
			name:     "info panel multiline",
			input:    "{info}\nThis is info content.\nSecond line.\n{info}",
			expected: "{panel:bgColor=#deebff}\nThis is info content.\nSecond line.\n{panel}",
		},
		{
			name:     "warning panel",
			input:    "{warning}Warning content{warning}",
			expected: "{panel:bgColor=#fffae6}Warning content{panel}",
		},
		{
			name:     "note panel",
			input:    "{note}Note content{note}",
			expected: "{panel:bgColor=#eae6ff}Note content{panel}",
		},
		{
			name:     "tip panel",
			input:    "{tip}Tip content{tip}",
			expected: "{panel:bgColor=#e3fcef}Tip content{panel}",
		},
		{
			name:     "error panel",
			input:    "{error}Error content{error}",
			expected: "{panel:bgColor=#ffebe6}Error content{panel}",
		},
		{
			name:     "success panel",
			input:    "{success}Success content{success}",
			expected: "{panel:bgColor=#e3fcef}Success content{panel}",
		},
		{
			name:     "info panel with title attribute",
			input:    "{info:title=Important}Content here{info}",
			expected: "{panel:bgColor=#deebff|title=Important}Content here{panel}",
		},
		{
			name:     "multiple panels",
			input:    "{info}Info{info}\n\n{warning}Warning{warning}",
			expected: "{panel:bgColor=#deebff}Info{panel}\n\n{panel:bgColor=#fffae6}Warning{panel}",
		},
		{
			name:     "generic panel unchanged",
			input:    "{panel}Panel content{panel}",
			expected: "{panel}Panel content{panel}",
		},
		{
			name:     "mixed content",
			input:    "Before\n{info}Info content{info}\nAfter",
			expected: "Before\n{panel:bgColor=#deebff}Info content{panel}\nAfter",
		},
		{
			name:     "unclosed panel unchanged",
			input:    "{info}Unclosed panel",
			expected: "{info}Unclosed panel",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, convertTypedPanels(tc.input))
		})
	}
}

func TestToJiraMD_TypedPanels(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "info panel converted",
			input:    "{info}Info content{info}",
			contains: "{panel:bgColor=#deebff}Info content{panel}",
		},
		{
			name:     "warning panel converted",
			input:    "{warning}Warning content{warning}",
			contains: "{panel:bgColor=#fffae6}Warning content{panel}",
		},
		{
			name:     "mixed markdown and panel",
			input:    "# Heading\n\n{info}Info content{info}\n\nParagraph",
			contains: "{panel:bgColor=#deebff}Info content{panel}",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToJiraMD(tc.input)
			assert.Contains(t, result, tc.contains)
		})
	}
}
