package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/carreter/discord-linkfixer-bot/pkg/fixer"
)

type ListFixersCommand struct {
	Store fixer.Store
}

func (c ListFixersCommand) ApplicationCommandTemplate() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "list-fixers",
		Description: "List all registered URL fixers for this server",
	}
}

func (c ListFixersCommand) Run(i *discordgo.InteractionCreate, opts map[string]any) (string, error) {
	fixers, err := c.Store.List(i.GuildID)
	if err != nil {
		return "", fmt.Errorf("could not list fixers: %w", err)
	}

	if len(fixers) == 0 {
		return "No fixers found!", nil
	}

	builder := strings.Builder{}
	builder.WriteString("Currently registered fixers:\n")
	for domain, fixer := range fixers {
		builder.WriteString(fmt.Sprintf("- `%v` → `%v`", domain, fixer.String()))
	}

	return builder.String(), nil
}
