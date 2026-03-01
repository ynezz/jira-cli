package md

import (
	"strings"
	"sync"
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
			name:     "nested typed panels convert inner and outer",
			input:    "{info}outer {warning}inner{warning} tail{info}",
			expected: "{panel:bgColor=#deebff}outer {panel:bgColor=#fffae6}inner{panel} tail{panel}",
		},
		{
			name:     "mismatched outer typed panels still convert nested valid pair",
			input:    "{info}outer {warning}inner{warning} tail{note}",
			expected: "{info}outer {panel:bgColor=#fffae6}inner{panel} tail{note}",
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

func TestToJiraMD_RegressionCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		input           string
		expectedContain string
		expectNoContain string
	}{
		{
			name: "markdown table containing links",
			input: `| Col1 | Col2 |
|---|---|
| [Link](url) | text |`,
			expectedContain: "||Col1||Col2||\n|[Link|url]|text|",
		},
		{
			name:            "horizontal rule",
			input:           "---",
			expectedContain: "----",
		},
		{
			name: "nested lists preserve depth",
			input: `- item1
    - subitem
        - subsubitem`,
			expectedContain: "* item1\n** subitem\n*** subsubitem",
		},
		{
			name: "nested lists with 2-space indentation preserve depth",
			input: `- Apple
- Banana
  - Yellow banana
  - Green banana
    - Very green banana
    - Slightly green banana
- Cherry`,
			expectedContain: "* Apple\n* Banana\n** Yellow banana\n** Green banana\n*** Very green banana\n*** Slightly green banana\n* Cherry",
		},
		{
			name:            "combined bold italic code mark nesting",
			input:           "**bold _italic `code`_**",
			expectedContain: "*bold _italic {{code}}_*",
		},
		{
			name: "all heading levels",
			input: `# H1
## H2
### H3
#### H4
##### H5
###### H6`,
			expectedContain: "h1. H1\nh2. H2\nh3. H3\nh4. H4\nh5. H5\nh6. H6",
		},
		{
			name:            "blockquote containing nested rich text",
			input:           "> **bold** and _italic_ in a quote",
			expectedContain: "{quote}\n*bold* and _italic_ in a quote",
		},
		{
			name:            "mismatched typed panel tags left unchanged",
			input:           "{info}content{warning}",
			expectedContain: "{info}content{warning}",
			expectNoContain: "{panel:bgColor=",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := ToJiraMD(tc.input)
			assert.Contains(t, result, tc.expectedContain)
			if tc.expectNoContain != "" {
				assert.NotContains(t, result, tc.expectNoContain)
			}
		})
	}
}

func TestToJiraMD_CodeBlockLanguageMapping(t *testing.T) {
	t.Parallel()

	t.Run("language-tagged fenced block uses code shorthand", func(t *testing.T) {
		t.Parallel()

		input := "```go\npackage main\nfunc main() {}\n```\n"
		result := ToJiraMD(input)

		assert.Contains(t, result, "{code:go}")
		assert.NotContains(t, result, "{code:language=go}")
		assert.NotContains(t, result, "{noformat}")
	})

	t.Run("plain fenced block uses noformat", func(t *testing.T) {
		t.Parallel()

		input := "```\nplain block, no language\n```\n"
		result := ToJiraMD(input)

		assert.Equal(t, 2, strings.Count(result, "{noformat}"))
		assert.NotContains(t, result, "{code}\nplain block, no language\n{code}")
	})

	t.Run("multiple fenced blocks keep boundaries", func(t *testing.T) {
		t.Parallel()

		input := "```go\npackage main\nfunc main() {}\n```\n\n```\nplain block, no language\n```\n"
		result := ToJiraMD(input)

		assert.Contains(t, result, "{code:go}\npackage main\nfunc main() {}\n{code}")
		assert.Contains(t, result, "{noformat}\nplain block, no language\n{noformat}")
		assert.NotContains(t, result, "{noformat}\n\n{noformat}\nplain block, no language")
	})

	t.Run("existing typed code macro is preserved", func(t *testing.T) {
		t.Parallel()

		input := "{code:go}\npackage main\nfunc main() {}\n{code}\n"
		result := ToJiraMD(input)

		assert.Contains(t, result, "{code:go}")
		assert.Contains(t, result, "\n{code}")
		assert.NotContains(t, result, "{noformat}")
	})

	t.Run("literal code macro inside plain fenced block is preserved", func(t *testing.T) {
		t.Parallel()

		input := "```\nline1\n{code}\nline3\n```\n"
		result := ToJiraMD(input)

		assert.Contains(t, result, "{noformat}\nline1\n{code}\nline3\n{noformat}")
		assert.Equal(t, 2, strings.Count(result, "{noformat}"))
	})

	t.Run("literal code macro inside language fenced block is preserved", func(t *testing.T) {
		t.Parallel()

		input := "```go\nline1\n{code}\nline3\n```\n"
		result := ToJiraMD(input)

		assert.Contains(t, result, "{noformat}\nline1\n{code}\nline3\n{noformat}")
		assert.NotContains(t, result, "{code:go}\nline1")
	})
}

func TestToJiraMD_ListIndentationNormalizerSkipsFencedCode(t *testing.T) {
	t.Parallel()

	input := "```\n  - keep spacing\n    - keep spacing too\n```\n\n- list\n  - nested\n"
	result := ToJiraMD(input)

	assert.Contains(t, result, "{noformat}\n  - keep spacing\n    - keep spacing too\n{noformat}")
	assert.Contains(t, result, "* list\n** nested")
}

func TestToJiraMDConcurrentRenderingProducesStableOutput(t *testing.T) {
	t.Parallel()

	const (
		iterations = 150
		workers    = 8
	)

	inputA := "- parent\n  - child\n"
	inputB := "- one\n"
	expectedA := ToJiraMD(inputA)
	expectedB := ToJiraMD(inputB)

	errs := make(chan string, iterations*workers*2)
	var wg sync.WaitGroup

	for range iterations {
		for range workers {
			wg.Add(2)

			go func() {
				defer wg.Done()
				if got := ToJiraMD(inputA); got != expectedA {
					errs <- "inputA"
				}
			}()

			go func() {
				defer wg.Done()
				if got := ToJiraMD(inputB); got != expectedB {
					errs <- "inputB"
				}
			}()
		}
	}

	wg.Wait()
	close(errs)

	for bad := range errs {
		t.Fatalf("non-deterministic output detected for %s under concurrent rendering", bad)
	}
}
