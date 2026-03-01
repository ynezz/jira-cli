# Plan: jira-cli Comprehensive ADF Test Coverage + Formatting Fixes

## Context

The `jira` CLI (ankitpokhrel/jira-cli, forked at `ynezz/jira-cli`) has confirmed formatting bugs in wiki table/link handling and several ADF rendering gaps. A separate code-block escaping issue has been reported historically, but should be treated as a hypothesis until reproduced on the current commit. This plan focuses on high-risk coverage and targeted fixes for confirmed gaps first.

**Repo**: `/data/projects/jira-cli` (branch: `ynezz/issue-attachments`)
**Installed**: v1.7.0-ynezz.2

---

## ADF Support Matrix (current state)

Three conversion paths in jira-cli:
- **Wiki→MD**: `pkg/md/jirawiki/parser.go` (Jira wiki markup → CommonMark)
- **MD→Wiki**: `pkg/md/md.go` + `github.com/kentaro-m/blackfriday-confluence` renderer (CommonMark → Jira wiki)
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
| panel (`{panel}`) | Y | Y | Y | SUPPORTED |
| typed panels (`{info}`/`{warning}`/`{note}`/`{tip}`/`{error}`) | — | partial | partial | **PARTIAL** (not parsed by Wiki→MD parser) |
| rule (hr) | — | Y | — | **PARTIAL** (MD→Wiki only) |
| expand | — | — | — | **NOT SUPPORTED** |
| nestedExpand | — | — | — | **NOT SUPPORTED** |
| mediaSingle/Group | — | — | partial | **PARTIAL** (only inner `media` nodes render as "[attachment]") |
| multiBodiedExtension | — | — | — | **NOT SUPPORTED** |

### Inline Nodes

| ADF Node | Wiki→MD | MD→Wiki | ADF→MD | Status |
|----------|---------|---------|--------|--------|
| text | Y | Y | Y | SUPPORTED |
| hardBreak | Y | Y | Y | SUPPORTED |
| mention | — | — | partial | **PARTIAL** (display text rendered, account ID not rendered) |
| emoji | — | — | partial | **PARTIAL** (output depends on `attrs.text`; no shortcode fallback) |
| inlineCard | — | — | partial | **PARTIAL** (renders a pin marker plus URL, no title/metadata) |
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

## Phase 1: Coverage expansion (target: +18 to +24 tests)

### `pkg/md/jirawiki/parser_test.go` (target: +10 to +14)

**Add to `TestTables`:**
1. Header cell containing Jira link: `||Name||[Link|https://example.com]||`
2. Body cell containing Jira link: `|cell|[Google|https://google.com]|`
3. Cell containing multiple links: `|[A|url1] and [B|url2]|text|`
4. Cell mixing bold+link syntax: `|*[Bold Link|url]*|`
5. Empty cell edge case: `||H1||H2||\n|||data|`

**Add to `TestParseReferenceLinks`:**
6. Link inside bold: `*[text|url]*`
7. Link inside italic: `_[text|url]_`
8. Multiple links in one line: `[A|url1] and [B|url2]`
9. Link with URL query+fragment: `[text|https://example.com/path?q=a&b=c#frag]`

**Add to `TestParseFencedCodeBlocks`:**
10. Code block containing wiki-looking content: `h1. Not a heading`
11. Code block containing table-looking content: `||not||a||table||`
12. Adjacent `{code}` / `{noformat}` blocks render independently
13. Code block containing blank lines and braces

**Add to `TestParseBlockQuote`:**
14. Blockquote containing list markers remains quoted content

Note: do not add horizontal-rule parsing tests in this file; current Wiki→MD parser has no `----` token support.

### `pkg/md/md_test.go` (target: +5 to +7)

Add focused `TestToJiraMD` (markdown → Jira wiki) cases:
15. Markdown table containing links (header/body)
16. Horizontal rule `---` → `----`
17. Nested lists preserve depth
18. Combined bold+italic+code mark nesting
19. All heading levels (1-6)
20. Blockquote containing nested rich text
21. Typed panel conversion (`{info}` / `{warning}`) through `ToJiraMD`

Gate any code-block escaping test on a reproducible failing fixture from current HEAD.

### `pkg/adf/adf_test.go` (target: +3 to +5)

Add focused ADF→Markdown fixtures:
22. Mention node renders display text; account ID omission is explicit
23. Inline card node renders URL text
24. Underline mark behavior (after implementation decision)
25. Code block preserves language attribute handling
26. Table cell text containing link-like content is not split incorrectly

Prefer small, targeted fixtures instead of extending the existing monolithic golden string only.

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

- `cd /data/projects/jira-cli && go test -race ./pkg/md/... ./pkg/adf/...`
- Manual: create ticket with all supported formatting elements
- Verify in Jira web UI at prplfoundationcloud.atlassian.net
