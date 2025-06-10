package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"

	bolt "go.etcd.io/bbolt"
)

const URLRegex = `https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`

func makeRegisterFixer(db *bolt.DB) func(s *discordgo.Session, m *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		matcher := ""
		replaceWith := ""
		options := i.ApplicationCommandData().Options
		for _, opt := range options {
			switch opt.Name {
			case "matcher":
				matcher = opt.StringValue()
			case "replace-with":
				replaceWith = opt.StringValue()
			}
		}

		if matcher == "" || replaceWith == "" {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Missing matcher or replace-with",
				},
			})
			if err != nil {
				log.Printf("could not respond to command: %v", err)
			}

			return
		}

		err := db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(i.GuildID))
			if b == nil {
				return fmt.Errorf("bucket %v does not exist", i.GuildID)
			}

			return b.Put([]byte(matcher), []byte(replaceWith))
		})
		if err != nil {
			log.Printf("could not register fixer, (matcher: %v, replace-with: %v): %v", matcher, replaceWith, err)
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Registered fixer with matcher `%v` and replace-with `%v`", matcher, replaceWith),
			},
		})
		if err != nil {
			log.Printf("could not respond to command: %v", err)
		}
	}
}

func makeListFixers(db *bolt.DB) func(s *discordgo.Session, m *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.ApplicationCommandData().Name != "list-fixers" {
			return
		}

		fixers := map[string]string{}

		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(i.GuildID))
			if b == nil {
				return fmt.Errorf("bucket %v does not exist", i.GuildID)
			}
			return b.ForEach(func(k, v []byte) error {
				fixers[string(k)] = string(v)
				return nil
			})
		})
		if err != nil {
			log.Printf("could not list fixers: %v", err)
		}

		responseBuilder := strings.Builder{}
		responseBuilder.WriteString("Currently registered fixers:\n")
		for k, v := range fixers {
			responseBuilder.WriteString(fmt.Sprintf("- `%v` to `%v`\n", k, v))
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: responseBuilder.String(),
			},
		})
		if err != nil {
			log.Printf("could not respond to command: %v", err)
		}
	}
}

func main() {
	authToken := flag.String("discord-auth-token", "", "Bot auth token")
	mappingsDbPath := flag.String("mappings-db", "./mappings.db", "Path to mappings database")
	flag.Parse()

	db, err := bolt.Open(*mappingsDbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err != nil {
		log.Printf("error creating mappings bucket: %v", err)
		return
	}

	discord, err := discordgo.New("Bot " + *authToken)
	if err != nil {
		log.Printf("error creating bot: %v", err)
		return
	}

	discord.AddHandler(makeUrlFixer(db))
	discord.AddHandler(makeRegisterFixer(db))
	discord.AddHandler(makeListFixers(db))
	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		for _, g := range r.Guilds {
			db.Update(func(tx *bolt.Tx) error {
				_, err := tx.CreateBucketIfNotExists([]byte(g.ID))
				return err
			})
		}

	})

	_, err = discord.ApplicationCommandCreate(discord.State.Application.ID, "", &discordgo.ApplicationCommand{
		Name:        "list-fixers",
		Description: "List URL fixers",
	})
	if err != nil {
		log.Printf("error creating command: %v", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	discord.Close()
}
