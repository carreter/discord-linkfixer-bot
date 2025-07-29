package linkfixerbot

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/carreter/discord-linkfixer-bot/pkg/fixer"
	"github.com/carreter/discord-linkfixer-bot/pkg/linkfixerbot/commands"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

var URLRegex = regexp.MustCompile(`https?:\/\/(?:www\.)?[-a-zA-Z0-9\._]+\/?[-a-zA-Z0-9()@:%_\+.~#?&\/=]+`)
var DomainRegex = regexp.MustCompile(`https?:\/\/(?:www\.)?([-a-zA-Z0-9\._]+)\/?`)

type LinkfixerBot struct {
	discord            *discordgo.Session
	commands           map[string]commands.Command
	registeredCommands []*discordgo.ApplicationCommand
	store              fixer.Store
}

func NewLinkfixerBot(authToken string, store fixer.Store) (*LinkfixerBot, error) {
	discord, err := discordgo.New("Bot " + authToken)
	if err != nil {
		return nil, fmt.Errorf("could not create discord session: %v", err)
	}

	// In this example, we only care about receiving message events.
	discord.Identify.Intents = discordgo.IntentsGuildMessages

	lb := &LinkfixerBot{
		discord: discord,
		commands: map[string]commands.Command{
			"replace-fixer":        commands.RegisterReplaceFixerCommand{Store: store},
			"register-csv-fixers":  commands.RegisterCsvFixersCommand{Store: store},
			"regexp-replace-fixer": commands.RegisterRegexpReplaceFixerCommand{Store: store},
			"prepend-fixer":        commands.RegisterPrependFixerCommand{Store: store},
			"list-fixers":          commands.ListFixersCommand{Store: store},
			"delete-fixer":         commands.DeleteFixerCommand{Store: store},
		},
		store: store,
	}

	lb.discord.AddHandler(lb.messageHandler)
	lb.discord.AddHandler(lb.interactionHandler)

	return lb, nil

}

func (lb *LinkfixerBot) interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		log.Warn("received interaction of unsupported type", "ID", i.ID, "type", i.Type)
		return
	}

	data := i.ApplicationCommandData()

	cmd, ok := lb.commands[data.Name]
	if !ok {
		log.Warn("received command with no registered handler", "interactionID", i.ID, "commandName", data.Name)
	}

	options := parseOptions(data.Options)
	response, err := cmd.Run(i, options)
	if err != nil {
		log.Error("could not run command", "commandName", data.Name, "options", options, "err", err)
		response = ":sob: Internal server error when running command"
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		log.Error("could not respond to command", "interactionID", i.ID, "err", err)
	}
}

func parseOptions(opts []*discordgo.ApplicationCommandInteractionDataOption) map[string]any {
	res := map[string]any{}
	for _, opt := range opts {
		res[opt.Name] = opt.Value
	}

	return res
}

func (lb *LinkfixerBot) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Skip messages we sent ourselves.
	if m.Author.ID == s.State.User.ID {
		return
	}

	mUrl := URLRegex.FindString(m.Content)
	if mUrl == "" { // Skip message without URLs.
		return
	}

	//  Chop off query params
	mUrl, _, _ = strings.Cut(mUrl, "?")

	domain := DomainRegex.FindStringSubmatch(mUrl)[1]
	f, err := lb.store.Get(m.GuildID, domain)
	if err != nil {
		log.Error("could not get domain from store", "domain", domain, "err", err)
		return
	}

	if f == nil {
		return
	}

	_, err = s.ChannelMessageSendReply(m.ChannelID, f.Fix(mUrl), m.Reference())
	if err != nil {
		log.Error("sending fixed link failed", "channelID", m.ChannelID, "messageID", m.ID)
		return
	}
}

func (lb *LinkfixerBot) Run(ctx context.Context) error {
	log.Info("opening connection to discord...")
	err := lb.discord.Open()
	if err != nil {
		fmt.Printf("error opening connection: %v\n", err)
		return err
	}
	log.Info("connected to discord!", "sessionID", lb.discord.State.SessionID)

	log.Info("creating commands...")
	err = lb.registerCommands()
	if err != nil {
		return fmt.Errorf("could not create commands: %w", err)
	}

	log.Info("bot ready")

	<-ctx.Done()

	log.Info("shutting down")
	err = lb.deleteCommands()
	if err != nil {
		log.Error("could not delete commands on shutdown", "err", err)
		return err
	}
	return nil
}

func (lb *LinkfixerBot) registerCommands() error {
	var commandTemplates []*discordgo.ApplicationCommand
	for _, command := range lb.commands {
		commandTemplates = append(commandTemplates, command.ApplicationCommandTemplate())
	}
	registeredCommands, err := lb.discord.ApplicationCommandBulkOverwrite(lb.discord.State.Application.ID, "", commandTemplates)
	if err != nil {
		return err
	}

	log.Info("successfully registered commands", "numRegistered", len(registeredCommands))
	lb.registeredCommands = registeredCommands
	return nil
}

func (lb *LinkfixerBot) deleteCommands() error {
	for _, command := range lb.registeredCommands {
		err := lb.discord.ApplicationCommandDelete(lb.discord.State.Application.ID, "", command.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
