package linkfixerbot

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

var URLRegex = regexp.MustCompile(`https?:\/\/(?:www\.)?[-a-zA-Z0-9\._]+\/?[-a-zA-Z0-9()@:%_\+.~#?&\/=]+`)
var DomainRegex = regexp.MustCompile(`https?:\/\/(?:www\.)?([-a-zA-Z0-9\._]+)\/?`)

type LinkfixerBot struct {
	discord  *discordgo.Session
	commands []*discordgo.ApplicationCommand
	store    Store
}

func NewLinkfixerBot(authToken string, store Store) (*LinkfixerBot, error) {
	discord, err := discordgo.New("Bot " + authToken)
	if err != nil {
		return nil, fmt.Errorf("could not create discord session: %v", err)
	}

	// In this example, we only care about receiving message events.
	discord.Identify.Intents = discordgo.IntentsGuildMessages

	lb := &LinkfixerBot{
		discord:  discord,
		commands: []*discordgo.ApplicationCommand{},
		store:    store,
	}

	lb.discord.AddHandler(lb.messageHandler)
	lb.discord.AddHandler(lb.registerReplaceFixerHandler)
	lb.discord.AddHandler(lb.listLinkfixersHandler)
	lb.discord.AddHandler(lb.registerPrependFixerHandler)

	return lb, nil

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

func (lb *LinkfixerBot) registerReplaceFixerHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name != "replace-fixer" {
		return
	}

	f := ReplaceFixer{}
	domain := ""
	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case "domain":
			domain = opt.StringValue()
		case "old":
			f.Old = opt.StringValue()
		case "new":
			f.New = opt.StringValue()
		}
	}

	if domain == "" || f.Old == "" {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "error: domain and old options must not be empty",
			},
		})
		if err != nil {
			log.Error("could not respond to command", "err", err)
		}
		return
	}

	err := lb.store.Put(i.GuildID, domain, f)
	if err != nil {
		log.Error("storing replace fixer failed", "domain", domain, "fixer", f, "err", err)
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "error: could not store fixer",
			},
		})
		if err != nil {
			log.Error("could not respond to command", "err", err)
			return
		}
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("added replace fixer for domain %v", domain),
		},
	})
	if err != nil {
		log.Error("could not respond to command", "err", err)
		return
	}
}

func (lb *LinkfixerBot) listLinkfixersHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name != "list-fixers" {
		return
	}

	fixers, err := lb.store.List(i.GuildID)
	if err != nil {
		log.Error("could not list fixers", "err", err)
	}

	response := ""
	if len(fixers) > 0 {
		builder := strings.Builder{}
		builder.WriteString("Currently registered fixers:\n")
		for domain, fixer := range fixers {
			builder.WriteString(fmt.Sprintf("- `%v` → `%+v`\n", domain, fixer))
		}
		response = builder.String()
	} else {
		response = "No currently registered fixers!"
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		log.Error("could not respond to command", "err", err)
	}
}

func (lb *LinkfixerBot) registerPrependFixerHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name != "prepend-fixer" {
		return
	}

	f := PrependFixer{}
	domain := ""
	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case "domain":
			domain = opt.StringValue()
		case "prefix":
			f.Prefix = opt.StringValue()
		}
	}

	err := lb.store.Put(i.GuildID, domain, f)
	if err != nil {
		log.Error("storing prepend fixer failed", "domain", domain, "fixer", f, "err", err)
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "error: could not store fixer",
			},
		})
		if err != nil {
			log.Error("could not respond to command", "err", err)
			return
		}
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("added prepend fixer for domain %v", domain),
		},
	})
	if err != nil {
		log.Error("could not respond to command", "err", err)
		return
	}
}

func (lb *LinkfixerBot) Run(ctx context.Context) error {

	// Open a websocket connection to Discord and begin listening.
	log.Info("opening connection to discord...")
	err := lb.discord.Open()
	if err != nil {
		fmt.Printf("error opening connection: %v\n", err)
		return err
	}
	log.Info("connected to discord!", "sessionID", lb.discord.State.SessionID)

	log.Info("creating commands...")
	err = lb.createCommands()
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

var commandTemplates = []*discordgo.ApplicationCommand{
	{
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
	},
	{
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
	},
	{
		Name:        "list-fixers",
		Description: "List all registered URL fixers for this server",
	},
}

func (lb *LinkfixerBot) createCommands() error {
	commands, err := lb.discord.ApplicationCommandBulkOverwrite(lb.discord.State.Application.ID, "", commandTemplates)
	if err != nil {
		return err
	}

	log.Info("successfully created commands", "len(commands)", len(commands))

	lb.commands = commands
	return nil
}

func (lb *LinkfixerBot) deleteCommands() error {
	for _, command := range lb.commands {
		err := lb.discord.ApplicationCommandDelete(lb.discord.State.Application.ID, "", command.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
