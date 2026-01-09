package attachment

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/jira-cli/internal/cmd/issue/attachment/add"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/issue/attachment/download"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/issue/attachment/list"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/issue/attachment/remove"
)

const helpText = `Attachment command helps you manage issue attachments. See available commands below.`

// NewCmdAttachment is an attachment command.
func NewCmdAttachment() *cobra.Command {
	cmd := cobra.Command{
		Use:     "attachment",
		Short:   "Manage issue attachments",
		Long:    helpText,
		Aliases: []string{"attachments", "att"},
		RunE:    attachment,
	}

	cmd.AddCommand(
		add.NewCmdAttachmentAdd(),
		list.NewCmdAttachmentList(),
		download.NewCmdAttachmentDownload(),
		remove.NewCmdAttachmentRemove(),
	)

	return &cmd
}

func attachment(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
