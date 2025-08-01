package commands

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/carreter/discord-linkfixer-bot/pkg/fixer"
)

type RegisterCsvFixersCommand struct {
	Store fixer.Store
}

func (c RegisterCsvFixersCommand) ApplicationCommandTemplate() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "register-csv-fixers",
		Description: "Register multiple URL fixers from a CSV string",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "csv",
				Description: "CSV file containing URL fixers",
				Required:    true,
			},
		},
	}
}

func (c RegisterCsvFixersCommand) Run(i *discordgo.InteractionCreate, opts map[string]any) (string, error) {
	attachments := i.ApplicationCommandData().Resolved.Attachments

	registeredCount := 0
	for _, attachment := range attachments {
		res, err := http.DefaultClient.Get(attachment.URL)
		if err != nil {
			return "", fmt.Errorf("could not download attachment: %w", err)
		}

		csv, _ := io.ReadAll(res.Body)
		res.Body.Close()

		fixers, err := parseFixers(string(csv))
		if err != nil {
			return "", fmt.Errorf("could not parse fixers: %w", err)
		}
		for domain, f := range fixers {
			err := c.Store.Put(i.GuildID, domain, f)
			if err != nil {
				return "", fmt.Errorf("storing fixer failed: %w", err)
			}
			registeredCount++
		}
	}

	return fmt.Sprintf("Successfully registered %v fixers.", registeredCount), nil
}

func parseFixers(csv string) (map[string]fixer.Fixer, error) {
	fixers := make(map[string]fixer.Fixer)
	for _, line := range strings.Split(csv, "\n") {
		cols := strings.Split(line, ",")
		var newFixer fixer.Fixer

		switch strings.TrimSpace(cols[0]) {
		case "prepend":
			if len(cols) != 3 {
				return nil, fmt.Errorf("invalid prepend fixer format (should be 'prepend,<domain>,<prefix>'): %v", line)
			}
			newFixer = fixer.PrependFixer{Prefix: strings.TrimSpace(cols[2])}
		case "replace":
			if len(cols) != 4 {
				return nil, fmt.Errorf("invalid prepend fixer format (should be 'replace,<domain>,<old>,<new>'): %v", line)
			}
			newFixer = fixer.ReplaceFixer{Old: strings.TrimSpace(cols[2]), New: strings.TrimSpace(cols[3])}
		default:
			return nil, fmt.Errorf("unknown fixer type: %s", cols[0])
		}

		domain := fixer.ExtractDomain(strings.TrimSpace(cols[1]))
		if domain == "" {
			return nil, fmt.Errorf("invalid domain: %s", cols[1])
		}
		fixers[domain] = newFixer
	}
	return fixers, nil
}
