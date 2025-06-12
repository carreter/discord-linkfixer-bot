package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	linkfixerbot "github.com/carreter/discord-linkfixer-bot/pkg"
	"github.com/charmbracelet/log"
	"go.etcd.io/bbolt"
)

func main() {
	authToken := flag.String("token", "", "Discord auth token")
	boltStore := flag.String("db", "./fixers.db", "path to database")

	flag.Parse()

	if *authToken == "" {
		log.Error("missing auth token")
		os.Exit(1)
	}

	db, err := bbolt.Open(*boltStore, 0600, &bbolt.Options{})
	if err != nil {
		log.Error("could not open bolt db", "path", *boltStore, "err", err)
		os.Exit(1)
	}

	bot, err := linkfixerbot.NewLinkfixerBot(*authToken, linkfixerbot.NewBoltStore(db))
	if err != nil {
		log.Error("could not create bot", "err", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context on os calls
	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		s := <-sc
		log.Info("received termination signal", "signal", s)
		cancel()
	}()

	err = bot.Run(ctx)
	if err != nil {
		log.Error("bot crashed :(", "err", err)
		os.Exit(1)
	}
}
