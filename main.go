package main

import (
	"context"
	"database/sql"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/markovic-dev/ajzak/internal/database"
	_ "modernc.org/sqlite"
)

func main() {
	godotenv.Load()
	token := os.Getenv("DISCORD_TOKEN")
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error Initialazing: %v", err)
	}
	db, err := sql.Open("sqlite", "ajzak.db")
	if err != nil {
		log.Fatalf("Error creating SQL connection: %v", err)
	}
	defer db.Close()
	queries := database.New(db)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if s.State.User.ID == m.Author.ID {
			return
		}
		if m.Content != ".provala" {
			return
		}
		msg, err := queries.GetLeastUsedProvale(context.Background())
		if err != nil {
			log.Printf("Unable to fetch responses from the database: %v", err)
			return
		}
		index := rand.IntN(len(msg))
		resp := msg[index]
		s.ChannelMessageSend(m.ChannelID, resp.Tekst)
		queries.IncrementProvalaUsage(context.Background(), resp.ID)
	})
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error establishing discord connection: %v", err)
	}
	defer dg.Close()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
}
