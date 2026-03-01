# Jira Cloud Manual QA Golden Issue Body

This fixture is used as a static probe issue description for manual
formatting verification.

## [TC-H1-H6] Heading Coverage

# H1 Heading
## H2 Heading
### H3 Heading
#### H4 Heading
##### H5 Heading
###### H6 Heading

## [TC-INLINE-STYLES] Inline Styles

This line includes **bold**, *italic*, ~~strikethrough~~, <u>underline</u>,
and `inline code`.

## [TC-TABLE-LINKS] Table Rendering and In-Cell Links

| Item | Docs Link | Query Link | Bare URL |
|---|---|---|---|
| jira-cli | [Project](https://github.com/ankitpokhrel/jira-cli) | [Search](https://example.com/search?q=jira-cli&sort=asc) | https://example.org/docs |
| RFC | [Spec](https://www.rfc-editor.org/rfc/rfc3986) | [Ref](https://example.net/path?alpha=1&beta=two) | https://atlassian.com/software/jira |

## [TC-LINK-VARIANTS] Link Variants

- Normal URL link: [Atlassian](https://www.atlassian.com/)
- URL with query string:
  [Issue Search](https://example.com/issues?project=TST&status=Open)
- Bare URL: https://developer.atlassian.com/cloud/jira/platform/

## [TC-CODEBLOCK-LANG] Fenced Code Block With Language

```go
package main

import "fmt"

func main() {
	fmt.Println("hello from jira-cli manual fixture")
}
```

## [TC-CODEBLOCK-PLAIN] Fenced Code Block Without Language

```
{info:title=Not a panel}
This must remain literal text in the code block.
{{macro}}
value := a|b|c
```

## [TC-ORDERED-LIST-3LVL] Ordered List, Nested 3 Levels

1. Level 0 item
   1. Level 1 item
      1. Level 2 item
2. Level 0 item two

## [TC-UNORDERED-LIST-3LVL] Unordered List, Nested 3 Levels

- Level 0 bullet
  - Level 1 bullet
    - Level 2 bullet
- Level 0 bullet two

## [TC-BLOCKQUOTE-MULTIPARA] Blockquote, Rich and Multi-Paragraph

> This is paragraph one with **bold text**, `inline code`, and a
> [link](https://example.com/quoted).
>
> This is paragraph two with additional details to verify paragraph
> separation in quoted text.

## [TC-PANELS-TYPED] Jira Typed Panels

{info:title=Info Panel}
Information panel body with `inline code`.
{info}

{warning:title=Warning Panel}
Warning panel body with a [link](https://example.com/warning).
{warning}

{note:title=Note Panel}
Note panel body.
{note}

{tip:title=Tip Panel}
Tip panel body.
{tip}

{error:title=Error Panel}
Error panel body.
{error}

{success:title=Success Panel}
Success panel body.
{success}

## [TC-HR] Horizontal Rule

---
