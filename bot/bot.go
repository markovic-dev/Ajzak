package bot

import (
	"database/sql"
	"fmt"
	"log"

	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/markovic-dev/ajzak/internal/database"
	_ "modernc.org/sqlite"
)

type Bot struct {
	Session *discordgo.Session
	Queries *database.Queries
}

func Start() {
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
		if !strings.HasPrefix(m.Content, ".") {
			return
		}
		bot := Bot{
			Session: s,
			Queries: queries,
		}
		switch {
		case m.Content == ".provala":
			bot.handleProvala(s, m)
		case m.Content == ".stats":
			bot.handleStats(s, m)
		case strings.HasPrefix(m.Content, ".dodaj "):
			bot.handleDodaj(s, m)
		default:
			bot.handleUnknown(s, m)
		}
	})
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error establishing discord connection: %v", err)
	}
	defer dg.Close()
	fmt.Println("Ajzak turned on succesfully!")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
	fmt.Println("Ajzak shutting down!")
}
