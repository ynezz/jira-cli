package list

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/internal/query"
	"github.com/ankitpokhrel/jira-cli/internal/view"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `List lists all attachments of an issue.`
	examples = `$ jira issue attachment list ISSUE-1

# List attachments in plain text format
$ jira issue attachment list ISSUE-1 --plain

# List attachments without headers
$ jira issue attachment list ISSUE-1 --plain --no-headers

# List specific columns
$ jira issue attachment list ISSUE-1 --columns id,filename,size`
)

// NewCmdAttachmentList is an attachment list command.
func NewCmdAttachmentList() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list ISSUE-KEY",
		Short:   "List issue attachments",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"ls"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1",
		},
		Run: listAttachments,
	}

	cmd.Flags().Bool("plain", false, "Display output in plain text format")
	cmd.Flags().Bool("no-headers", false, "Don't print headers in plain text output")
	cmd.Flags().String("columns", "", fmt.Sprintf("Comma separated list of columns to display.\nAccepts: %s", strings.Join(view.ValidAttachmentColumns(), ", ")))

	return &cmd
}

func listAttachments(cmd *cobra.Command, args []string) {
	params := parseArgsAndFlags(args, cmd.Flags())

	if params.issueKey == "" {
		cmdutil.Failed("Issue key is required")
	}

	client := api.DefaultClient(params.debug)

	attachments, err := func() (interface{}, error) {
		s := cmdutil.Info("Fetching attachments")
		defer s.Stop()

		return client.GetAttachments(params.issueKey)
	}()
	cmdutil.ExitIfError(err)

	v := view.AttachmentList{
		Server:   viper.GetString("server"),
		IssueKey: params.issueKey,
		Data:     attachments.([]*jira.Attachment),
		Display: view.DisplayFormat{
			Plain:     params.plain,
			NoHeaders: params.noHeaders,
			Columns:   params.columns,
			Timezone:  viper.GetString("timezone"),
		},
		Refresh: func() { listAttachments(cmd, args) },
	}

	cmdutil.ExitIfError(v.RenderInTable())
}

type listParams struct {
	issueKey  string
	plain     bool
	noHeaders bool
	columns   []string
	debug     bool
}

func parseArgsAndFlags(args []string, flags query.FlagParser) *listParams {
	var issueKey string

	if len(args) >= 1 {
		issueKey = cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])
	}

	debug, err := flags.GetBool("debug")
	cmdutil.ExitIfError(err)

	plain, err := flags.GetBool("plain")
	cmdutil.ExitIfError(err)

	noHeaders, err := flags.GetBool("no-headers")
	cmdutil.ExitIfError(err)

	columnsStr, err := flags.GetString("columns")
	cmdutil.ExitIfError(err)

	var columns []string
	if columnsStr != "" {
		columns = strings.Split(columnsStr, ",")
	}

	return &listParams{
		issueKey:  issueKey,
		plain:     plain,
		noHeaders: noHeaders,
		columns:   columns,
		debug:     debug,
	}
}
