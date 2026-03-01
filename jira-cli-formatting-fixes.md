# Plan: jira-cli Comprehensive ADF Test Coverage + Formatting Fixes

## Context

The `jira` CLI (ankitpokhrel/jira-cli, forked at `ynezz/jira-cli`) has formatting bugs in its wiki markup conversion that break links in tables, escape code block content, and miss some ADF marks. This plan adds comprehensive test coverage for ALL supported ADF formatting elements and fixes the known bugs.

**Repo**: `/var/home/ynezz/dev/go/jira-cli` (branch: `ynezz/issue-attachments`)
**Installed**: v1.7.0-ynezz.2

---

## ADF Support Matrix (current state)

Three conversion paths in jira-cli:
- **Wiki→MD**: `pkg/md/jirawiki/parser.go` (Jira wiki markup → CommonMark)
- **MD→Wiki**: `vendor/github.com/kentaro-m/blackfriday-confluence/confluence.go` (CommonMark → Jira wiki)
- **ADF→MD**: `pkg/adf/markdown.go` + `jiramarkdown.go` (ADF JSON → CommonMark for terminal display)

### Block Nodes

| ADF Node | Wiki→MD | MD→Wiki | ADF→MD | Status |
|----------|---------|---------|--------|--------|
| paragraph | Y | Y | Y | SUPPORTED |
| heading (1-6) | Y | Y | Y | SUPPORTED |
| bulletList | Y | Y | Y | SUPPORTED |
| orderedList | Y | Y | Y | SUPPORTED |
| listItem | Y | Y | Y | SUPPORTED |
| blockquote | Y | Y | Y | SUPPORTED |
| codeBlock+language | Y | Y | Y | SUPPORTED |
| table/row/cell/header | Y | Y | Y | SUPPORTED (BUG: links in cells) |
| panel (5 types) | Y | Y | Y | SUPPORTED |
| rule (hr) | — | Y | — | **PARTIAL** (MD→Wiki only) |
| expand | — | — | — | **NOT SUPPORTED** |
| nestedExpand | — | — | — | **NOT SUPPORTED** |
| mediaSingle/Group | — | — | partial | **PARTIAL** (renders as "[attachment]") |
| multiBodiedExtension | — | — | — | **NOT SUPPORTED** |

### Inline Nodes

| ADF Node | Wiki→MD | MD→Wiki | ADF→MD | Status |
|----------|---------|---------|--------|--------|
| text | Y | Y | Y | SUPPORTED |
| hardBreak | Y | Y | Y | SUPPORTED |
| mention | — | — | partial | **PARTIAL** (ID lost) |
| emoji | — | — | partial | **PARTIAL** (renders as space) |
| inlineCard | — | — | partial | **PARTIAL** (URL lost) |
| date | — | — | — | **NOT SUPPORTED** |
| status | — | — | — | **NOT SUPPORTED** |
| mediaInline | — | — | — | **NOT SUPPORTED** |

### Marks

| ADF Mark | Wiki→MD | MD→Wiki | ADF→MD | Status |
|----------|---------|---------|--------|--------|
| strong (bold) | Y | Y | Y | SUPPORTED |
| em (italic) | Y | Y | Y | SUPPORTED |
| code (inline) | Y | Y | Y | SUPPORTED |
| link | Y | Y | Y | SUPPORTED |
| strike | Y | Y | Y | SUPPORTED |
| underline | — | — | — | **NOT SUPPORTED** (in test data but no renderer) |
| textColor | — | — | — | **NOT SUPPORTED** |
| backgroundColor | — | — | — | **NOT SUPPORTED** |
| subsup | — | — | — | **NOT SUPPORTED** |
| border | — | — | — | **NOT SUPPORTED** |

---

## Phase 1: Comprehensive test cases (~39 tests)

### `pkg/md/jirawiki/parser_test.go` (~20 tests)

**Add to `TestTables`:**
1. Table header with link: `||Name||[Link|https://example.com]||`
2. Table cell with link: `|cell|[Google|https://google.com]|`
3. Table cell with multiple links: `|[A|url1] and [B|url2]|text|`
4. Table cell with bold+link: `|*[Bold Link|url]*|`
5. Empty table cells: `||H1||H2||\n|||data|`

**Add to `TestParseFencedCodeBlocks`:**
6. Code block with `()[]{}*_-` characters (verify no escaping)
7. Code block with wiki-looking content: `h1. Not a heading`
8. Code block with table-looking content: `||not||a||table||`
9. Adjacent code blocks with different languages
10. Code block with blank lines inside

**Add to `TestParseReferenceLinks`:**
11. Link inside bold: `*[text|url]*`
12. Link inside italic: `_[text|url]_`
13. Multiple links in one line: `[A|url1] and [B|url2]`
14. Link with special chars in URL: `[text|https://example.com/path?q=a&b=c#frag]`

**Add to `TestParseListTags`:**
15. List item with link: `* [Google|https://google.com]`
16. List item with code: `* some {{code}} here`
17. Nested list with mixed bullet/ordered: `* item\n## ordered sub`
18. List item with bold+italic: `* *bold* _italic_`

**Add to `TestParsePanels`:**
19. Panel with list inside: `{info}\n* item1\n* item2\n{info}`
20. Panel with code block inside: `{warning}\n{code}...\n{code}\n{warning}`
21. Panel with link: `{note}See [here|url]{note}`
22. All 5 panel types: `{info}`, `{warning}`, `{note}`, `{tip}`, `{error}`

### `pkg/md/md_test.go` (~7 tests)

Add to `TestToJiraMD` (markdown → Jira wiki):
23. Fenced code block with special chars → verify `{code}` output has no escaping
24. Table with links → verify `||header||` output
25. Horizontal rule `---` → verify `----` output
26. Nested lists → verify proper formatting
27. Combined bold+italic+code → verify proper nesting
28. All heading levels (1-6)
29. Blockquote with rich content

### `pkg/adf/adf_test.go` (~7 tests)

Add ADF→Markdown test cases:
30. ADF with underline mark → should render (currently missing handler)
31. ADF with mention preserving ID
32. ADF with inlineCard preserving URL
33. ADF with all 5 panel types
34. ADF with nested lists (bullet inside ordered)
35. ADF with table containing links in cells
36. ADF with code block + language attribute

### New: `TestParseBlockQuoteNested` and `TestParseHorizontalRule`
37. Blockquote with list: `{quote}\n* item\n{quote}`
38. Blockquote with bold/italic content
39. `----` horizontal rule

---

## Phase 2: Fix known bugs

### Bug 1: Links inside table cells

**Files**: `pkg/md/jirawiki/parser.go` (lines 282-290, 482-501)

**Root cause**: `tokenize()` claims the entire table line as one token. `handleTable()` does `strings.ReplaceAll(line, "||", "|")`, destroying `[text|url]` link syntax.

**Fix**:
- New helper `splitTableCells(line string, isHeader bool) []string`
- Tracks `[` / `]` bracket depth; only treats `|` as separator at depth 0
- Convert `[text|url]` → `[text](url)` within each cell
- Update `handleTable()` to use the new splitter

### Bug 2: Code block escaping (investigate)

**Files**: `pkg/md/md.go` (line 60), `vendor/github.com/kentaro-m/blackfriday-confluence/confluence.go`

The confluence renderer's `esc()` escapes `()[]{}*_-+^~![`. The `IgnoreMacroEscaping` flag only exempts `{`.

**Fix** (if confirmed): Patch vendored renderer to skip `esc()` inside `CodeBlock` nodes, or post-process in `ToJiraMD()` to un-escape `{code}...{code}` content.

### Bug 3: Underline mark not rendered (ADF→MD)

**File**: `pkg/adf/markdown.go`

Test data includes underline marks but `MarkdownTranslator` has no handler for `MarkUnderline`.

**Fix**: Add underline handling following strike pattern (lines 151-152).

---

## Phase 3: Build & release

1. Branch `ynezz/formatting-fixes` from `ynezz/issue-attachments`
2. Add all test cases first (TDD)
3. Implement fixes for bugs 1-3
4. Run full test suite: `go test -race ./...`
5. Manual test: create ticket with all formatting elements, verify in Jira web UI
6. Tag `v1.7.0-ynezz.3`, build via GoReleaser, install

---

## Files to modify

| File | Changes |
|------|---------|
| `pkg/md/jirawiki/parser.go` | Add `splitTableCells()`, update `handleTable()` |
| `pkg/md/jirawiki/parser_test.go` | Add ~20 test cases across existing test functions |
| `pkg/md/md.go` | Possible post-processing for code block escaping |
| `pkg/md/md_test.go` | Add ~7 test cases for MD→Wiki direction |
| `pkg/adf/markdown.go` | Add underline mark handling |
| `pkg/adf/adf_test.go` | Add ~7 test cases for ADF→MD direction |
| `vendor/.../confluence.go` | Possible patch for code block escaping |

## Verification

- `cd /var/home/ynezz/dev/go/jira-cli && go test -race ./pkg/md/... ./pkg/adf/...`
- Manual: create ticket with all supported formatting elements
- Verify in Jira web UI at prplfoundationcloud.atlassian.net
