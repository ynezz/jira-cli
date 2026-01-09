package view

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/atotto/clipboard"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
	"github.com/ankitpokhrel/jira-cli/pkg/tui"
)

const attachmentHelpText = `[default]ATTACHMENT LIST ACTIONS
---------------------------

* [yellow]← → ↑ ↓ / j, k, h, l[default] to navigate
* [yellow]ENTER[default] to download selected attachment
* [yellow]c[default] to copy attachment URL to clipboard
* [yellow]CTRL + k[default] to copy attachment ID to clipboard
* [yellow]CTRL + r / F5[default] to refresh the list
* [yellow]q / ESC / CTRL + c[default] to quit
* [yellow]?[default] to view this help`

// AttachmentList is a list view for attachments.
type AttachmentList struct {
	Server   string
	IssueKey string
	Data     []*jira.Attachment
	Display  DisplayFormat
	Refresh  tui.RefreshFunc
}

// ValidAttachmentColumns returns valid columns for attachment list.
func ValidAttachmentColumns() []string {
	return []string{
		fieldID,
		fieldFilename,
		fieldSize,
		fieldAuthor,
		fieldCreated,
		fieldMimeType,
	}
}

// RenderInTable renders the list in table view.
func (al *AttachmentList) RenderInTable() error {
	if al.Display.Plain || tui.IsDumbTerminal() || tui.IsNotTTY() {
		w := tabwriter.NewWriter(os.Stdout, 0, tabWidth, 1, '\t', 0)
		return al.renderPlain(w)
	}

	data := al.tableData()
	view := tui.NewTable(
		tui.WithFixedColumns(al.Display.FixedColumns),
		tui.WithTableStyle(al.Display.TableStyle),
		tui.WithTableFooterText(fmt.Sprintf("Showing %d attachments", len(al.Data))),
		tui.WithTableHelpText(attachmentHelpText),
		tui.WithDownloadFunc(al.downloadAttachment()),
		tui.WithCopyFunc(copyAttachmentURL(al.Data)),
		tui.WithCopyKeyFunc(copyAttachmentID()),
		tui.WithRefreshFunc(al.Refresh),
	)

	return view.Paint(data)
}

// renderPlain renders the attachments in plain view.
func (al *AttachmentList) renderPlain(w io.Writer) error {
	return renderPlain(w, al.tableData(), "\t")
}

func (al *AttachmentList) validColumnsMap() map[string]struct{} {
	columns := ValidAttachmentColumns()
	out := make(map[string]struct{}, len(columns))

	for _, c := range columns {
		out[c] = struct{}{}
	}

	return out
}

func (al *AttachmentList) tableHeader() []string {
	if len(al.Display.Columns) == 0 {
		return ValidAttachmentColumns()
	}

	var headers []string

	columnsMap := al.validColumnsMap()
	for _, c := range al.Display.Columns {
		c = strings.ToUpper(c)
		if _, ok := columnsMap[c]; ok {
			headers = append(headers, c)
		}
	}

	return headers
}

func (al *AttachmentList) tableData() tui.TableData {
	var data tui.TableData

	headers := al.tableHeader()

	// Fallback to all valid columns if no valid columns were specified
	if len(headers) == 0 {
		headers = ValidAttachmentColumns()
	}

	if !(al.Display.Plain && al.Display.NoHeaders) {
		data = append(data, headers)
	}

	for _, a := range al.Data {
		data = append(data, al.assignColumns(headers, a))
	}

	return data
}

func (al *AttachmentList) assignColumns(columns []string, attachment *jira.Attachment) []string {
	var bucket []string

	for _, column := range columns {
		switch column {
		case fieldID:
			bucket = append(bucket, attachment.ID)
		case fieldFilename:
			bucket = append(bucket, attachment.Filename)
		case fieldSize:
			bucket = append(bucket, formatSize(attachment.Size))
		case fieldAuthor:
			bucket = append(bucket, attachment.Author.DisplayName)
		case fieldCreated:
			bucket = append(bucket, formatDateTime(attachment.Created, jira.RFC3339MilliLayout, al.Display.Timezone))
		case fieldMimeType:
			bucket = append(bucket, attachment.MimeType)
		}
	}

	return bucket
}

// formatSize formats byte size to human readable format.
func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// downloadAttachment returns a DownloadFunc that downloads the selected attachment.
func (al *AttachmentList) downloadAttachment() tui.DownloadFunc {
	return func(r, _ int, d interface{}) func() {
		if r == 0 { // Skip header row
			return nil
		}

		data := d.(tui.TableData)
		attachmentID := data.Get(r, data.GetIndex(fieldID))
		filename := data.Get(r, data.GetIndex(fieldFilename))

		// Return function to execute while screen is suspended
		return func() {
			client := api.DefaultClient(false)

			s := cmdutil.Info(fmt.Sprintf("Downloading %q", filename))
			err := client.DownloadAttachment(attachmentID, filename)
			s.Stop()

			if err != nil {
				fmt.Printf("\nFailed to download: %v\n", err)
			} else {
				cmdutil.Success("Downloaded %q", filename)
			}

			// Pause to let user see the result
			fmt.Print("Press ENTER to continue...")
			_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
		}
	}
}

// copyAttachmentURL returns a CopyFunc that copies the attachment content URL.
func copyAttachmentURL(attachments []*jira.Attachment) tui.CopyFunc {
	return func(r, _ int, _ interface{}) {
		if r == 0 || r > len(attachments) {
			return
		}
		_ = clipboard.WriteAll(attachments[r-1].Content)
	}
}

// copyAttachmentID returns a CopyKeyFunc that copies the attachment ID.
func copyAttachmentID() tui.CopyKeyFunc {
	return func(r, _ int, d interface{}) {
		if r == 0 {
			return
		}
		data := d.(tui.TableData)
		id := data.Get(r, data.GetIndex(fieldID))
		_ = clipboard.WriteAll(id)
	}
}
