package linkfixerbot

import (
	"context"
	"fmt"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	
	bolt "go.etcd.io/bbolt"
)

var URLRegex = regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
var DomainRegex = regexp.MustCompile(`https?:\/\/(?:www\.)?(-a-zA-Z0-9\._)\/`)

type LinkfixerBot struct {
	discord  *discordgo.Session
	commands []*discordgo.ApplicationCommand
	db *bolt.DB
}

func NewLinkfixerBot(authToken string, store Store) (*LinkfixerBot, error) {
	discord, err := discordgo.New("Bot " + authToken)
	if err != nil {
		return nil, fmt.Errorf("could not create discord session: %v", err)
	}

	// In this example, we only care about receiving message events.
	discord.Identify.Intents = discordgo.IntentsGuildMessages

	return &LinkfixerBot{
		discord:  discord,
		commands: []*discordgo.ApplicationCommand{},
		store:    store,
	}, nil

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
	f, err := lb.store.Get(domain)
	if err != nil {
		log.Error("could not get domain from store", "domain", domain)
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

	err := lb.store.Put(domain, f)
	if err != nil {
		log.Error("storing replace fixer failed", "domain", domain, "fixer", f)
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
			Content: fmt.Sprintf("added fixer for domain %v", domain),
		},
	})
	if err != nil {
		log.Error("could not respond to command", "err", err)
		return
	}
}

func (lb *LinkfixerBot) Run(ctx context.Context) {
	lb.discord.AddHandler(lb.messageHandler)
	lb.discord.AddHandler(lb.registerReplaceFixerHandler)

	// Open a websocket connection to Discord and begin listening.
	err := lb.discord.Open()
	if err != nil {
		fmt.Printf("error opening connection: %v\n", err)
		return
	}

}

var commands = []*discordgo.ApplicationCommand{
	cmd, err := lb.discord.ApplicationCommandCreate(lb.discord.State.Application.ID, "", &discordgo.ApplicationCommand{
		Name:        "register-fixer",
		Description: "Register a URL fixer",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "matcher",
				Description: "regex to match against URLs",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "replace-with",
				Description: "replace URLs with this",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	})

}

func (lb *LinkfixerBot) createCommands() {

	if err != nil {
		log.Printf("error creating command: %v", err)
		return
	}
}
