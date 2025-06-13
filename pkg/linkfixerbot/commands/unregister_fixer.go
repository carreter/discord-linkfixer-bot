package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/carreter/discord-linkfixer-bot/pkg/fixer"
)

type DeleteFixerCommand struct {
	Store fixer.Store
}

func (c DeleteFixerCommand) ApplicationCommandTemplate() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "delete-fixer",
		Description: "Delete a fixer for a domain",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "domain",
				Description: "Domain of the fixer to delete",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}
}

func (c DeleteFixerCommand) Run(i *discordgo.InteractionCreate, opts map[string]any) (string, error) {
	domain := opts["domain"].(string)

	err := c.Store.Delete(i.GuildID, domain)
	if err != nil {
		return "", fmt.Errorf("deleting fixer failed: %w", err)
	}

	return fmt.Sprintf("Successfully deleted fixer for domain `%v`", domain), nil
}
