package commands

import (
	"github.com/bwmarrin/discordgo"
)

type Command interface {
	ApplicationCommandTemplate() *discordgo.ApplicationCommand
	Run(i *discordgo.InteractionCreate, opts map[string]any) (string, error)
}
