# Manual Test Plan (Jira Cloud Formatting Verification)

This plan is the reusable manual QA procedure distilled from `bd-10c` and
`bd-3cn`.

Use it as the pre-release verification gate for Markdown/Wiki/ADF rendering
changes.

## Canonical Inputs

- Probe issue body (static golden Markdown):
  [`tests/manual/golden-jira-issue.md`](golden-jira-issue.md)
- Probe case index and checklist mapping (static golden JSON):
  [`tests/manual/golden-jira-issue.json`](golden-jira-issue.json)

Treat these files as canonical test assets. Any additions to manual
rendering coverage must update both files in the same change.

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

## Create Probe Issue From Golden Fixture

Create one probe issue and set its description to the exact contents of
`tests/manual/golden-jira-issue.md`.

Example flow:

1. Copy the fixture body exactly (no manual edits).
2. Create issue with the candidate binary in target project.
3. Record the resulting issue key for evidence.

## Execution Steps

1. Create probe issue from `golden-jira-issue.md`.
2. Open the issue in Jira web UI.
3. Validate forward rendering checklist.
4. Run reverse rendering checks:
   - `jira issue view <KEY> --plain`
   - `jira issue view <KEY> --raw`
5. Capture evidence (ticket key, pass/fail matrix, notes).
6. If any failure appears, create a new bead with
   `discovered-from:<manual-test-bead-id>`.

## Case Mapping

Use stable case IDs from `golden-jira-issue.json` when recording failures.

| Checkpoint | Case IDs |
|---|---|
| Headings H1..H6 hierarchy | `TC-H1-H6` |
| Tables and links in cells | `TC-TABLE-LINKS` |
| Text emphasis + inline code | `TC-INLINE-STYLES` |
| URL, query URL, bare URL | `TC-LINK-VARIANTS` |
| Fenced code with language | `TC-CODEBLOCK-LANG` |
| Fenced code without language + special chars/macros | `TC-CODEBLOCK-PLAIN` |
| Ordered list depth 0/1/2 | `TC-ORDERED-LIST-3LVL` |
| Unordered list depth 0/1/2 | `TC-UNORDERED-LIST-3LVL` |
| Rich multi-paragraph blockquote | `TC-BLOCKQUOTE-MULTIPARA` |
| Typed panels info/warning/note/error/success (+ tip alias) | `TC-PANELS-TYPED` |
| Horizontal rule separator | `TC-HR` |

## Forward Rendering Checklist (Web UI)

- [ ] `TC-TABLE-LINKS`: Tables render with correct boundaries and aligned columns
- [ ] `TC-TABLE-LINKS`: Links inside table cells are preserved and clickable
- [ ] `TC-H1-H6`: Headings H1..H6 show expected hierarchy
- [ ] `TC-INLINE-STYLES`: Bold/italic/strike/underline/inline code render distinctly
- [ ] `TC-CODEBLOCK-LANG`: Code block language-tagged sections preserve language
- [ ] `TC-CODEBLOCK-PLAIN`: Code block without language does not default incorrectly
- [ ] `TC-ORDERED-LIST-3LVL`: Ordered list depth 0/1/2 renders correctly
- [ ] `TC-UNORDERED-LIST-3LVL`: Unordered list depth 0/1/2 renders correctly
- [ ] `TC-BLOCKQUOTE-MULTIPARA`: Blockquote rich text/multi-paragraph renders correctly
- [ ] `TC-PANELS-TYPED`: Panels render by expected cloud types; tip/success share the green success-style panel and must remain distinguishable by title/content
- [ ] `TC-HR`: Horizontal rule renders as separator

## Reverse Rendering Checklist (CLI)

- [ ] `TC-TABLE-LINKS`: `--plain` output keeps table boundary clean
- [ ] `TC-PANELS-TYPED`: `--plain` output shows panel indicators; tip/success may both appear as `[SUCCESS]` and should be distinguished by panel title/content
- [ ] `TC-BLOCKQUOTE-MULTIPARA`: Multi-paragraph blockquote remains readable
- [ ] `TC-CODEBLOCK-LANG`, `TC-CODEBLOCK-PLAIN`: Code blocks and inline code are readable
- [ ] `TC-LINK-VARIANTS`: No unintended escaping/corruption in rendered URLs/text
- [ ] `TC-H1-H6`: `--raw` ADF shape shows expected heading node hierarchy

## Result Template

```markdown
## Manual Verification Result

- Candidate build: <commit/tag/version>
- Jira project: <project key>
- Probe issue: <KEY>
- Date: <YYYY-MM-DD>
- Golden assets: tests/manual/golden-jira-issue.md,
  tests/manual/golden-jira-issue.json

### Forward Path
- [ ] PASS
- Notes:

### Reverse Path
- [ ] PASS
- Notes:

### Failed Case IDs
- <TC-...> (if any)

### Follow-up Beads
- <bead-id> (if any)
```
