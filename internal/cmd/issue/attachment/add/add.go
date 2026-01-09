package add

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/internal/query"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `Add uploads file(s) as attachments to an issue.`
	examples = `$ jira issue attachment add ISSUE-1 --file /path/to/file.txt

# Upload multiple files
$ jira issue attachment add ISSUE-1 --file file1.txt --file file2.png

# Open issue in browser after uploading
$ jira issue attachment add ISSUE-1 --file /path/to/file.txt --web`
)

// NewCmdAttachmentAdd is an attachment add command.
func NewCmdAttachmentAdd() *cobra.Command {
	cmd := cobra.Command{
		Use:     "add ISSUE-KEY --file FILE [--file FILE...]",
		Short:   "Add attachment(s) to an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"upload"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1",
		},
		Run: addAttachment,
	}

	cmd.Flags().StringArrayP("file", "f", []string{}, "Path to file(s) to upload (required, can be specified multiple times)")
	cmd.Flags().Bool("web", false, "Open issue in web browser after adding attachment")

	return &cmd
}

func addAttachment(cmd *cobra.Command, args []string) {
	params := parseArgsAndFlags(args, cmd.Flags())

	if params.issueKey == "" {
		cmdutil.Failed("Issue key is required")
	}

	if len(params.files) == 0 {
		cmdutil.Failed("At least one file is required (use --file flag)")
	}

	// Validate all files exist before uploading
	for _, filePath := range params.files {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			cmdutil.Failed("File not found: %s", filePath)
		}
	}

	client := api.DefaultClient(params.debug)

	for _, filePath := range params.files {
		attachments, err := func() (interface{}, error) {
			s := cmdutil.Info(fmt.Sprintf("Uploading %q", filepath.Base(filePath)))
			defer s.Stop()

			return client.AddAttachment(params.issueKey, filePath)
		}()
		cmdutil.ExitIfError(err)

		att := attachments.([]*jira.Attachment)
		if len(att) > 0 {
			cmdutil.Success("Attachment %q added to issue %q (ID: %s)", att[0].Filename, params.issueKey, att[0].ID)
		}
	}

	server := viper.GetString("server")
	fmt.Printf("%s\n", cmdutil.GenerateServerBrowseURL(server, params.issueKey))

	if web, _ := cmd.Flags().GetBool("web"); web {
		err := cmdutil.Navigate(server, params.issueKey)
		cmdutil.ExitIfError(err)
	}
}

type addParams struct {
	issueKey string
	files    []string
	debug    bool
}

func parseArgsAndFlags(args []string, flags query.FlagParser) *addParams {
	var issueKey string

	if len(args) >= 1 {
		issueKey = cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])
	}

	files, err := flags.GetStringArray("file")
	cmdutil.ExitIfError(err)

	debug, err := flags.GetBool("debug")
	cmdutil.ExitIfError(err)

	return &addParams{
		issueKey: issueKey,
		files:    files,
		debug:    debug,
	}
}
