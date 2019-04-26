package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/bwmarrin/discordgo"
	"google.golang.org/api/option"
)

const (
	collection = "stream"
	docID      = "stream-key"
)

func dochieURL(url string) error {
	opt := option.WithCredentialsFile("./dochie-firebase-sdk.json")
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	update := []firestore.Update{
		firestore.Update{
			Path:  "url",
			Value: url,
		},
	}

	_, err = client.Collection(collection).Doc(docID).Update(ctx, update)
	if err != nil {
		return err
	}

	return nil
}

func dochieHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	url := m.Content
	if !strings.HasPrefix(url, "https://") {
		s.ChannelMessageSend(m.ChannelID, "Please for stream URL")
		return
	} else {
		res, err := http.Get(url)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
		if res.StatusCode != 200 {
			fmt.Fprint(os.Stderr, errors.New("bad url"))
			return
		}
	}

	if err := dochieURL(url); err != nil {
		fmt.Fprint(os.Stderr, err)
	} else {
		s.ChannelMessageSend(m.ChannelID, "Done!")
	}

}

func main() {
	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	dg.AddHandler(dochieHandler)

	if err := dg.Open(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	defer dg.Close()

	<-make(chan struct{})
}
