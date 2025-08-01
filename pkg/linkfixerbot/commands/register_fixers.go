package commands

import (
	"fmt"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/carreter/discord-linkfixer-bot/pkg/fixer"
)

type RegisterReplaceFixerCommand struct {
	Store fixer.Store
}

func (c RegisterReplaceFixerCommand) ApplicationCommandTemplate() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "replace-fixer",
		Description: "Register a URL fixer that replaces one substring in a URL with another",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "domain",
				Description: "Domain this fixer will apply to",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "old",
				Description: "Substring to replace",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "new",
				Description: "What to replace old substring with",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}
}

func (c RegisterReplaceFixerCommand) Run(i *discordgo.InteractionCreate, opts map[string]any) (string, error) {
	domain := fixer.ExtractDomain(opts["domain"].(string))
	f := fixer.ReplaceFixer{
		Old: opts["old"].(string),
		New: opts["new"].(string),
	}

	err := c.Store.Put(i.GuildID, domain, f)
	if err != nil {
		return "", fmt.Errorf("storing prepend fixer failed: %w", err)
	}

	return fmt.Sprintf("Successfully registered replace fixer `%v` for domain `%v`", f.String(), domain), nil
}

type RegisterRegexpReplaceFixerCommand struct {
	Store fixer.Store
}

func (c RegisterRegexpReplaceFixerCommand) ApplicationCommandTemplate() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "regexp-replace-fixer",
		Description: "Register a URL fixer that replaces one substring in a URL with another",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "domain",
				Description: "Domain this fixer will apply to",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "pattern",
				Description: "Regular expression to match against URL",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "replacement",
				Description: "Replacement string, reference capture groups with $x",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}
}

func (c RegisterRegexpReplaceFixerCommand) Run(i *discordgo.InteractionCreate, opts map[string]any) (string, error) {
	domain := fixer.ExtractDomain(opts["domain"].(string))
	f := fixer.RegexpReplaceFixer{
		Pattern:     opts["pattern"].(string),
		Replacement: opts["replacement"].(string),
	}

	_, err := regexp.Compile(f.Pattern)
	if err != nil {
		return "could not compile regular expression `%v`", nil
	}

	err = c.Store.Put(i.GuildID, domain, f)
	if err != nil {
		return "", fmt.Errorf("storing prepend fixer failed: %w", err)
	}

	return fmt.Sprintf("Successfully registered fixer `%v` for domain `%v`", f.String(), domain), nil
}

type RegisterPrependFixerCommand struct {
	Store fixer.Store
}

func (c RegisterPrependFixerCommand) ApplicationCommandTemplate() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "prepend-fixer",
		Description: "Register a URL fixer that prepends a string to a URL",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "domain",
				Description: "Domain this fixer will apply to",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "prefix",
				Description: "Prefix to prepend",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}
}

func (c RegisterPrependFixerCommand) Run(i *discordgo.InteractionCreate, opts map[string]any) (string, error) {
	domain := fixer.ExtractDomain(opts["domain"].(string))
	f := fixer.PrependFixer{
		Prefix: opts["prefix"].(string),
	}

	err := c.Store.Put(i.GuildID, domain, f)
	if err != nil {
		return "", fmt.Errorf("storing prepend fixer failed: %w", err)
	}

	return fmt.Sprintf("Successfully registered fixer `%v` for domain `%v`", f.String(), domain), nil
}
