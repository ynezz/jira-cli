package remove

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/internal/query"
)

const (
	helpText = `Remove deletes an attachment from an issue.`
	examples = `$ jira issue attachment remove 10000

# Remove attachment without confirmation
$ jira issue attachment remove 10000 --force`
)

// NewCmdAttachmentRemove is an attachment remove command.
func NewCmdAttachmentRemove() *cobra.Command {
	cmd := cobra.Command{
		Use:     "remove ATTACHMENT-ID",
		Short:   "Remove an attachment",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"rm", "delete", "del"},
		Annotations: map[string]string{
			"help:args": "ATTACHMENT-ID\tAttachment ID to remove",
		},
		Run: removeAttachment,
	}

	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	return &cmd
}

func removeAttachment(cmd *cobra.Command, args []string) {
	params := parseArgsAndFlags(args, cmd.Flags())

	if params.attachmentID == "" {
		cmdutil.Failed("Attachment ID is required")
	}

	if !params.force {
		var confirm bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Are you sure you want to delete attachment %q?", params.attachmentID),
		}
		err := survey.AskOne(prompt, &confirm)
		cmdutil.ExitIfError(err)

		if !confirm {
			cmdutil.Failed("Action aborted")
		}
	}

	client := api.DefaultClient(params.debug)

	err := func() error {
		s := cmdutil.Info(fmt.Sprintf("Removing attachment %q", params.attachmentID))
		defer s.Stop()

		return client.DeleteAttachment(params.attachmentID)
	}()
	cmdutil.ExitIfError(err)

	cmdutil.Success("Attachment %q removed successfully", params.attachmentID)
}

type removeParams struct {
	attachmentID string
	force        bool
	debug        bool
}

func parseArgsAndFlags(args []string, flags query.FlagParser) *removeParams {
	var attachmentID string

	if len(args) >= 1 {
		attachmentID = args[0]
	}

	force, err := flags.GetBool("force")
	cmdutil.ExitIfError(err)

	debug, err := flags.GetBool("debug")
	cmdutil.ExitIfError(err)

	return &removeParams{
		attachmentID: attachmentID,
		force:        force,
		debug:        debug,
	}
}
