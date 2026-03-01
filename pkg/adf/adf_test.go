package adf

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestADF(t *testing.T) {
	data, err := os.ReadFile("./testdata/md.json")
	assert.NoError(t, err)

	var adf ADF
	err = json.Unmarshal(data, &adf)
	assert.NoError(t, err)

	tr := NewTranslator(&adf, NewMarkdownTranslator())

	expected := "# H1\n## H2\n1. Some text\n\n2. Some more text\n\n\n\n> Blockquote text\n\n\nInline Node 📍 https://antiklabs.atlassian.net/wiki/spaces/ANK/pages/124234/hello-world \n\nImplement epic browser\n\n---\nPanel paragraph\n\n---\n @Person A \n\n---\n **Strong** Paragraph 1\n\nParagraph 2\n\n---\n **Bold Text** \n\n _Italic Text_ \n\n <u>Prefix: Underlined Text</u> \n\n `Prefix: Inline Code Block` \n\n -Prefix: Strikethrough text- \n\n [Link](https://ankit.pl) \n\n- Prefix: Unordered list item 1\n\t- Next\n\t\t- Another\n\t\t\t- New level\n- Unordered list item 2\n- Unordered list item 3\n1. Ordered list item 1\n2. Ordered list item 2\n3. Ordered list item 3\n\t1. nested\n\t\t1. second level\n\t\t\t1. third level\n\t\t\t\t1. fourth level\n\n **Table Header 1**  |  **Table Header 2**  |  **Table Header 3** \n--- | --- | ---\nTable row 1 column 1 | Table row 1 column 2 | Table row 1 column 3\nTable row 2 column 1 | Table row 2 column 2 | Table row 2 column 3\n```go\npackage main\n\nimport (\n\t\"fmt\"\n)\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n```\n\n **Table Header 1**  |  **Table Header 2**  |  **Table Header 3**  |  **Table Header 4**  |  **Table Header 5** \n--- | --- | --- | --- | ---\nTable row 1 column 1 | Table row 2 column 1 | Table row 3 column 1 | Table row 4 column 1 | Table row 5 column 1\nTable row 1 column 2 | Table row 2 column 2 | Table row 3 column 2 | Table row 4 column 2 | Table row 5 column 2\nTable row 1 column 2 | Table row 2 column 3 | Table row 3 column 3 | Table row 4 column 3 | Table row 5 column 3\n"
	expected = strings.Replace(
		expected,
		"Table row 2 column 1 | Table row 2 column 2 | Table row 2 column 3\n```go",
		"Table row 2 column 1 | Table row 2 column 2 | Table row 2 column 3\n\n```go",
		1,
	)
	expected += "\n"
	assert.Equal(t, expected, tr.Translate())
}

func TestADFReplaceAll(t *testing.T) {
	data, err := os.ReadFile("./testdata/md.json")
	assert.NoError(t, err)

	var adf ADF
	err = json.Unmarshal(data, &adf)
	assert.NoError(t, err)

	adf.ReplaceAll("Prefix:", "Replaced:")

	dump, err := json.Marshal(adf)
	assert.NoError(t, err)

	assert.False(t, strings.Contains(string(dump), "Prefix:"))
	assert.True(t, strings.Contains(string(dump), "Replaced:"))
}

func TestMarkdownTranslatorMentionRendersDisplayText(t *testing.T) {
	doc := &ADF{
		Version: 1,
		DocType: "doc",
		Content: []*Node{
			{
				NodeType: NodeParagraph,
				Content: []*Node{
					{
						NodeType:   InlineNodeMention,
						Attributes: map[string]any{"id": "abc123", "text": "John Doe"},
					},
				},
			},
		},
	}

	out := newMarkdownTranslator(doc)
	assert.Contains(t, out, "@John Doe")
	assert.NotContains(t, out, "abc123")
}

func TestMarkdownTranslatorInlineCardRendersURL(t *testing.T) {
	doc := &ADF{
		Version: 1,
		DocType: "doc",
		Content: []*Node{
			{
				NodeType: NodeParagraph,
				Content: []*Node{
					{
						NodeType:   InlineNodeCard,
						Attributes: map[string]any{"url": "https://example.com"},
					},
				},
			},
		},
	}

	out := newMarkdownTranslator(doc)
	assert.Contains(t, out, "https://example.com")
}

func TestMarkdownTranslatorUnderlineMarkRendersHTMLTag(t *testing.T) {
	doc := &ADF{
		Version: 1,
		DocType: "doc",
		Content: []*Node{
			{
				NodeType: NodeParagraph,
				Content: []*Node{
					{
						NodeType: ChildNodeText,
						NodeValue: NodeValue{
							Text: "Underlined Text",
							Marks: []MarkNode{
								{MarkType: NodeType("underline")},
							},
						},
					},
				},
			},
		},
	}

	out := newMarkdownTranslator(doc)
	assert.Contains(t, out, "<u>Underlined Text</u>")
}

func TestMarkdownTranslatorCodeBlockPreservesLanguage(t *testing.T) {
	doc := &ADF{
		Version: 1,
		DocType: "doc",
		Content: []*Node{
			{
				NodeType:   NodeCodeBlock,
				Attributes: map[string]any{"language": "python"},
				Content: []*Node{
					{
						NodeType: ChildNodeText,
						NodeValue: NodeValue{
							Text: `print("hi")`,
						},
					},
				},
			},
		},
	}

	out := newMarkdownTranslator(doc)
	assert.Contains(t, out, "```python")
	assert.Contains(t, out, `print("hi")`)
}

func TestMarkdownTranslatorTableCellPipeCurrentBehavior(t *testing.T) {
	doc := &ADF{
		Version: 1,
		DocType: "doc",
		Content: []*Node{
			{
				NodeType: NodeTable,
				Content: []*Node{
					{
						NodeType: ChildNodeTableRow,
						Content: []*Node{
							{
								NodeType: ChildNodeTableHeader,
								Content: []*Node{
									{
										NodeType: ChildNodeText,
										NodeValue: NodeValue{
											Text: "H1",
										},
									},
								},
							},
							{
								NodeType: ChildNodeTableHeader,
								Content: []*Node{
									{
										NodeType: ChildNodeText,
										NodeValue: NodeValue{
											Text: "H2",
										},
									},
								},
							},
						},
					},
					{
						NodeType: ChildNodeTableRow,
						Content: []*Node{
							{
								NodeType: ChildNodeTableCell,
								Content: []*Node{
									{
										NodeType: ChildNodeText,
										NodeValue: NodeValue{
											Text: "a|b",
										},
									},
								},
							},
							{
								NodeType: ChildNodeTableCell,
								Content: []*Node{
									{
										NodeType: ChildNodeText,
										NodeValue: NodeValue{
											Text: "value",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	out := newMarkdownTranslator(doc)
	assert.Contains(t, out, "a|b")
}

func TestMarkdownTranslatorTableThenParagraphHasBlankLineBoundary(t *testing.T) {
	doc := &ADF{
		Version: 1,
		DocType: "doc",
		Content: []*Node{
			{
				NodeType: NodeTable,
				Content: []*Node{
					{
						NodeType: ChildNodeTableRow,
						Content: []*Node{
							{
								NodeType: ChildNodeTableHeader,
								Content: []*Node{
									{
										NodeType: ChildNodeText,
										NodeValue: NodeValue{
											Text: "Col A",
										},
									},
								},
							},
							{
								NodeType: ChildNodeTableHeader,
								Content: []*Node{
									{
										NodeType: ChildNodeText,
										NodeValue: NodeValue{
											Text: "Col B",
										},
									},
								},
							},
						},
					},
					{
						NodeType: ChildNodeTableRow,
						Content: []*Node{
							{
								NodeType: ChildNodeTableCell,
								Content: []*Node{
									{
										NodeType: ChildNodeText,
										NodeValue: NodeValue{
											Text: "left",
										},
									},
								},
							},
							{
								NodeType: ChildNodeTableCell,
								Content: []*Node{
									{
										NodeType: ChildNodeText,
										NodeValue: NodeValue{
											Text: "right",
										},
									},
								},
							},
						},
					},
				},
			},
			{
				NodeType: NodeParagraph,
				Content: []*Node{
					{
						NodeType: ChildNodeText,
						NodeValue: NodeValue{
							Text: "Ordered list:",
						},
					},
				},
			},
		},
	}

	out := newMarkdownTranslator(doc)
	assert.Contains(t, out, "left | right\n\nOrdered list:")
}

func TestJiraMarkdownTranslatorUnderlineMarkUsesJiraSyntax(t *testing.T) {
	doc := &ADF{
		Version: 1,
		DocType: "doc",
		Content: []*Node{
			{
				NodeType: NodeParagraph,
				Content: []*Node{
					{
						NodeType: ChildNodeText,
						NodeValue: NodeValue{
							Text: "Underlined Text",
							Marks: []MarkNode{
								{MarkType: MarkUnderline},
							},
						},
					},
				},
			},
		},
	}

	out := newJiraMarkdownTranslator(doc)
	assert.Contains(t, out, "+Underlined Text+")
}

func newMarkdownTranslator(doc *ADF) string {
	tr := NewTranslator(doc, NewMarkdownTranslator())
	return tr.Translate()
}

func newJiraMarkdownTranslator(doc *ADF) string {
	tr := NewTranslator(doc, NewJiraMarkdownTranslator())
	return tr.Translate()
}
