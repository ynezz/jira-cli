# Release Runbook

End-to-end procedure for shipping a jira-cli release: manual QA, tagging,
GoReleaser build, and beads bookkeeping.

## Prerequisites

- Access to a Jira Cloud instance with a test project (e.g. `TST`).
- A working `jira-cli` build from the branch under test.
- `br` (beads_rust) available for issue tracking.
- `ubs` available for bug scanning.
- GoReleaser v2 installed (or CI will run it on tag push).

---

## Phase 1: Pre-release Manual QA

### 1.1 Create the QA bead

```bash
br create "Manual QA for v<VERSION>" -t task -p 1 --json
br update <qa-bead> --status in_progress --json
```

Historical examples: `bd-258` (v1.7.0-ynezz.4), `bd-10c` / `bd-3cn`
(earlier formatting rounds).

### 1.2 Build the candidate binary

```bash
cd /data/projects/jira-cli
go build -o /tmp/jira-cli-manual ./cmd/jira
/tmp/jira-cli-manual version
```

Record the commit hash and version string for the QA report.

### 1.3 Execute the manual test plan

Follow [`tests/manual/manual-test-plan.md`](../../tests/manual/manual-test-plan.md)
end-to-end. Do not invent ad-hoc variants.

Two directions must be verified:

1. **Forward path:** `Markdown -> Jira Wiki -> Jira ADF -> Jira Web UI`
2. **Reverse path:** `Jira ADF -> jira issue view --plain/--raw`

Use the golden fixtures as canonical inputs:

- [`tests/manual/golden-jira-issue.md`](../../tests/manual/golden-jira-issue.md) —
  static issue-description Markdown
- [`tests/manual/golden-jira-issue.json`](../../tests/manual/golden-jira-issue.json) —
  stable case/check mapping for structured reporting

### 1.4 Record results

Use stable case IDs from `golden-jira-issue.json` when recording pass/fail:

| Checkpoint | Case ID |
|---|---|
| Headings H1..H6 hierarchy | `TC-H1-H6` |
| Tables and links in cells | `TC-TABLE-LINKS` |
| Text emphasis + inline code | `TC-INLINE-STYLES` |
| URL, query URL, bare URL | `TC-LINK-VARIANTS` |
| Fenced code with language | `TC-CODEBLOCK-LANG` |
| Fenced code without language | `TC-CODEBLOCK-PLAIN` |
| Ordered list depth 0/1/2 | `TC-ORDERED-LIST-3LVL` |
| Unordered list depth 0/1/2 | `TC-UNORDERED-LIST-3LVL` |
| Rich multi-paragraph blockquote | `TC-BLOCKQUOTE-MULTIPARA` |
| Typed panels (info/warning/note/error/success/tip) | `TC-PANELS-TYPED` |
| Horizontal rule separator | `TC-HR` |

Minimum completion artifacts:

- Candidate commit/tag tested
- Jira project + probe issue key
- Forward-path result
- Reverse-path result
- Follow-up bead IDs for any failures

### 1.5 Handle failures

For every failing case, create a follow-up bead linked to the QA bead:

```bash
br create "Fix <TC-ID>: <description>" -t bug -p 1 \
  --deps discovered-from:<qa-bead> --json
```

Fix the issues, rebuild the candidate (step 1.2), and re-run the failing
checks. Do not close the QA bead until all blocking cases pass.

### 1.6 Close the QA bead

```bash
br close <qa-bead> --reason "All cases pass on <commit>" --json
```

---

## Phase 2: Tag the Release

### 2.1 Create the release bead

```bash
br create "Release v<VERSION>" -t task -p 1 \
  --deps <qa-bead> --json
br update <release-bead> --status in_progress --json
```

Historical example: `bd-157` (v1.7.0-ynezz.3 / ynezz.4), `bd-9vc`.

### 2.2 Verify all blockers are closed

```bash
br show <release-bead> --json   # Check deps are all closed
br list --status=open --json    # Confirm no surprise blockers
```

### 2.3 Sync branch state

```bash
git pull --rebase
```

### 2.4 Quality gates

Both must pass before tagging:

```bash
go test -race ./...
ubs --diff --only=golang
```

### 2.5 Create annotated tag

```bash
git tag -a v<VERSION> -m "Release v<VERSION>"
```

---

## Phase 3: Create the Release

### 3.1 Push tag (triggers CI GoReleaser)

```bash
git push origin v<VERSION>
```

This triggers the GitHub Actions workflow
([`.github/workflows/release.yml`](../../.github/workflows/release.yml))
which runs GoReleaser v2.

The CI workflow:

- Checks out with full history (`fetch-depth: 0`)
- Sets up Go ^1.24.1
- Runs `goreleaser release --clean`
- On `workflow_dispatch` with `snapshot=true`, runs in snapshot mode and
  uploads artifacts instead of creating a release

### 3.2 Local GoReleaser build (alternative)

For testing or when CI is unavailable:

```bash
goreleaser release --clean
```

Or snapshot mode (no actual release):

```bash
goreleaser release --snapshot --clean
```

GoReleaser configuration ([`.goreleaser.yml`](../../.goreleaser.yml)):

- **Binary:** `bin/jira` from `./cmd/jira`
- **Platforms:** macOS (amd64, arm64), Linux (386, arm, amd64, arm64),
  Windows (amd64)
- **Ldflags:** Version, GitCommit, SourceDateEpoch injected via
  `internal/version`
- **Archives:** `.tar.gz` for macOS/Linux, `.zip` for Windows
- **Checksums:** SHA-256 in `checksums.txt`
- **Homebrew:** Formula generated (skip_upload: true — manual publish)
- **Release:** Draft mode, prerelease auto-detected from tag

### 3.3 Post release summary on release bead

Include in the bead close reason or a comment:

- Version/tag
- Artifact/build output reference
- Link to manual QA result (probe issue key, pass/fail summary)

### 3.4 Close release bead

```bash
br close <release-bead> --reason "v<VERSION> released, QA passed" --json
```

### 3.5 Sync beads

```bash
br sync --flush-only
git add .beads/ && git commit -s -m "beads: sync after v<VERSION> release"
git push
```

---

## Quick Reference

| Phase | Gate | Bead type | Example |
|---|---|---|---|
| 1. Manual QA | All case IDs pass (forward + reverse) | task | `bd-258` |
| 2. Tag | `go test -race` + `ubs` pass, blockers closed | task | `bd-157` |
| 3. Release | CI GoReleaser succeeds, summary posted | task | `bd-9vc` |

## See Also

- [`tests/manual/manual-test-plan.md`](../../tests/manual/manual-test-plan.md) —
  canonical manual QA procedure
- [`tests/manual/golden-jira-issue.md`](../../tests/manual/golden-jira-issue.md) —
  golden Markdown fixture
- [`tests/manual/golden-jira-issue.json`](../../tests/manual/golden-jira-issue.json) —
  case/check mapping
- [`.goreleaser.yml`](../../.goreleaser.yml) — GoReleaser configuration
- [`.github/workflows/release.yml`](../../.github/workflows/release.yml) —
  CI release workflow
