# Manual Test Plan (Jira Cloud Formatting Verification)

This plan is the reusable manual QA procedure distilled from `bd-10c` and
`bd-3cn`.

Use it as the pre-release verification gate for Markdown/Wiki/ADF rendering
changes.

## Goal

Verify both directions end-to-end:

1. Forward path: `Markdown -> Jira Wiki -> Jira ADF -> Jira Web UI`
2. Reverse path: `Jira ADF -> CLI issue view output`

## Preconditions

- You can access the target Jira Cloud instance.
- You have a working `jira-cli` build from the branch under test.
- You have credentials/config for the target project (for example `TST`).

## Build And Install Candidate

```bash
cd /data/projects/jira-cli
go build -o /tmp/jira-cli-manual ./cmd/jira
/tmp/jira-cli-manual version
```

Optional local install:

```bash
go install ./cmd/jira
jira version
```

## Canonical Test Fixture

Create one probe issue whose description contains all of:

- Headings H1..H6
- Tables (including links inside cells)
- Bold, italic, strikethrough, underline, inline code
- Links: normal URL, URL with query string, bare URL
- Code blocks:
  - fenced with language (for example `go`)
  - fenced without language
  - code containing special characters/macros
- Ordered list nested 3 levels
- Unordered list nested 3 levels
- Blockquote with rich text and multi-paragraph content
- Typed panels: `{info}`, `{warning}`, `{note}`, `{tip}`, `{error}`,
  `{success}`
- Horizontal rule (`---`)

## Execution Steps

1. Create probe issue using the candidate CLI.
2. Open the issue in Jira web UI.
3. Validate forward rendering checklist.
4. Run reverse rendering checks:
   - `jira issue view <KEY> --plain`
   - `jira issue view <KEY> --raw`
5. Capture evidence (ticket key, pass/fail matrix, notes).
6. If any failure appears, create a new bead with
   `discovered-from:<manual-test-bead-id>`.

## Forward Rendering Checklist (Web UI)

- [ ] Tables render with correct boundaries and aligned columns
- [ ] Links inside table cells are preserved and clickable
- [ ] Headings H1..H6 show expected hierarchy
- [ ] Bold/italic/strike/underline/inline code render distinctly
- [ ] Code block language-tagged sections preserve language
- [ ] Code block without language does not default incorrectly
- [ ] Ordered list depth 0/1/2 renders correctly
- [ ] Unordered list depth 0/1/2 renders correctly
- [ ] Blockquote with rich text/multi-paragraph renders correctly
- [ ] Panels render and are distinguishable by type
- [ ] Horizontal rule renders as separator

## Reverse Rendering Checklist (CLI)

- [ ] `--plain` output keeps table boundary clean
- [ ] `--plain` output shows panel type indicators
- [ ] Blockquote multi-paragraph content remains readable
- [ ] Code blocks and inline code are readable and bounded correctly
- [ ] No unintended escaping/corruption in rendered text
- [ ] `--raw` ADF shape matches expectations for key fields

## Result Template

```markdown
## Manual Verification Result

- Candidate build: <commit/tag/version>
- Jira project: <project key>
- Probe issue: <KEY>
- Date: <YYYY-MM-DD>

### Forward Path
- [ ] PASS
- Notes:

### Reverse Path
- [ ] PASS
- Notes:

### Follow-up Beads
- <bead-id> (if any)
```
