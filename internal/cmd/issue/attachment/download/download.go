package download

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/internal/query"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `Download downloads attachment(s) from an issue.`
	examples = `$ jira issue attachment download ISSUE-1 10000

# Download to specific file
$ jira issue attachment download ISSUE-1 10000 --output /path/to/file.txt

# Download all attachments from an issue
$ jira issue attachment download ISSUE-1 --all

# Download all attachments to a specific directory
$ jira issue attachment download ISSUE-1 --all --output-dir /path/to/dir

# Overwrite existing files
$ jira issue attachment download ISSUE-1 --all --overwrite`
)

// NewCmdAttachmentDownload is an attachment download command.
func NewCmdAttachmentDownload() *cobra.Command {
	cmd := cobra.Command{
		Use:     "download ISSUE-KEY [ATTACHMENT-ID]",
		Short:   "Download attachment(s) from an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"dl", "get"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1\n" +
				"ATTACHMENT-ID\tAttachment ID to download (optional with --all)",
		},
		Run: downloadAttachment,
	}

	cmd.Flags().StringP("output", "o", "", "Output file path (for single attachment)")
	cmd.Flags().StringP("output-dir", "d", ".", "Output directory (for --all)")
	cmd.Flags().Bool("all", false, "Download all attachments from the issue")
	cmd.Flags().Bool("overwrite", false, "Overwrite existing files")

	return &cmd
}

func downloadAttachment(cmd *cobra.Command, args []string) {
	params := parseArgsAndFlags(args, cmd.Flags())

	if params.issueKey == "" {
		cmdutil.Failed("Issue key is required")
	}

	client := api.DefaultClient(params.debug)

	if params.all {
		downloadAllAttachments(client, params)
	} else {
		downloadSingleAttachment(client, params)
	}
}

func downloadSingleAttachment(client *jira.Client, params *downloadParams) {
	attachmentID := params.attachmentID

	// If no attachment ID provided, list and prompt for selection
	if attachmentID == "" {
		attachments, err := client.GetAttachments(params.issueKey)
		cmdutil.ExitIfError(err)

		if len(attachments) == 0 {
			cmdutil.Failed("No attachments found for issue %q", params.issueKey)
		}

		options := make([]string, len(attachments))
		idMap := make(map[string]string)
		for i, att := range attachments {
			options[i] = fmt.Sprintf("%s - %s (%s)", att.ID, att.Filename, formatSize(att.Size))
			idMap[options[i]] = att.ID
		}

		var selected string
		prompt := &survey.Select{
			Message: "Select attachment to download:",
			Options: options,
		}
		err = survey.AskOne(prompt, &selected)
		cmdutil.ExitIfError(err)

		attachmentID = idMap[selected]
	}

	// Get attachment metadata if we need the filename
	outputPath := params.output
	if outputPath == "" {
		attachments, err := client.GetAttachments(params.issueKey)
		cmdutil.ExitIfError(err)

		for _, att := range attachments {
			if att.ID == attachmentID {
				outputPath = att.Filename
				break
			}
		}
		if outputPath == "" {
			outputPath = fmt.Sprintf("attachment-%s", attachmentID)
		}
	}

	// Check if file exists
	if !params.overwrite {
		if _, err := os.Stat(outputPath); err == nil {
			cmdutil.Failed("File %q already exists (use --overwrite to replace)", outputPath)
		}
	}

	err := func() error {
		s := cmdutil.Info(fmt.Sprintf("Downloading attachment %q", attachmentID))
		defer s.Stop()

		return client.DownloadAttachment(attachmentID, outputPath)
	}()
	cmdutil.ExitIfError(err)

	cmdutil.Success("Attachment downloaded to %q", outputPath)
}

func downloadAllAttachments(client *jira.Client, params *downloadParams) {
	attachments, err := func() (interface{}, error) {
		s := cmdutil.Info("Fetching attachments")
		defer s.Stop()

		return client.GetAttachments(params.issueKey)
	}()
	cmdutil.ExitIfError(err)

	atts := attachments.([]*jira.Attachment)
	if len(atts) == 0 {
		cmdutil.Failed("No attachments found for issue %q", params.issueKey)
	}

	// Ensure output directory exists
	if params.outputDir != "." {
		err := os.MkdirAll(params.outputDir, os.ModePerm)
		cmdutil.ExitIfError(err)
	}

	for _, att := range atts {
		outputPath := filepath.Join(params.outputDir, att.Filename)

		// Check if file exists
		if !params.overwrite {
			if _, err := os.Stat(outputPath); err == nil {
				fmt.Printf("Skipping %q (file exists, use --overwrite to replace)\n", att.Filename)
				continue
			}
		}

		err := func() error {
			s := cmdutil.Info(fmt.Sprintf("Downloading %q", att.Filename))
			defer s.Stop()

			return client.DownloadAttachment(att.ID, outputPath)
		}()
		if err != nil {
			fmt.Printf("Failed to download %q: %v\n", att.Filename, err)
			continue
		}

		cmdutil.Success("Downloaded %q", att.Filename)
	}
}

type downloadParams struct {
	issueKey     string
	attachmentID string
	output       string
	outputDir    string
	all          bool
	overwrite    bool
	debug        bool
}

func parseArgsAndFlags(args []string, flags query.FlagParser) *downloadParams {
	var issueKey, attachmentID string

	nargs := len(args)
	if nargs >= 1 {
		issueKey = cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])
	}
	if nargs >= 2 {
		attachmentID = args[1]
	}

	output, err := flags.GetString("output")
	cmdutil.ExitIfError(err)

	outputDir, err := flags.GetString("output-dir")
	cmdutil.ExitIfError(err)

	all, err := flags.GetBool("all")
	cmdutil.ExitIfError(err)

	overwrite, err := flags.GetBool("overwrite")
	cmdutil.ExitIfError(err)

	debug, err := flags.GetBool("debug")
	cmdutil.ExitIfError(err)

	return &downloadParams{
		issueKey:     issueKey,
		attachmentID: attachmentID,
		output:       output,
		outputDir:    outputDir,
		all:          all,
		overwrite:    overwrite,
		debug:        debug,
	}
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
