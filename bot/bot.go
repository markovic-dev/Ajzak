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
	Session        *discordgo.Session
	Queries        *database.Queries
	VoiceConn      *discordgo.VoiceConnection
	NodeServiceURL string
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
	b := &Bot{
		Session:        dg,
		Queries:        queries,
		NodeServiceURL: "http://localhost:3000",
	}
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates | discordgo.IntentMessageContent
	dg.AddHandler(b.HandleInteraction)
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if s.State.User.ID == m.Author.ID {
			return
		}
		if !strings.HasPrefix(m.Content, ".") {
			return
		}
		switch {
		case m.Content == ".provala":
			b.handleProvala(s, m)
		case m.Content == ".stats":
			b.handleStats(s, m)
		case strings.HasPrefix(m.Content, ".dodaj "):
			b.handleDodaj(s, m)
		case m.Content == ".leave":
			b.handleLeave(s, m)
		case m.Content == ".join":
			b.handleJoin(s, m)
		case m.Content == ".skip":
			b.handleSkip(s, m)
		case m.Content == ".queue":
			b.handleQueue(s, m)
		case m.Content == ".stop":
			b.handleStop(s, m)
		case strings.HasPrefix(m.Content, ".pusti") || strings.HasPrefix(m.Content, ".play"):
			args := strings.Fields(m.Content)
			if len(args) > 1 {
				b.handlePlay(s, m, args[1:])
			} else {
				b.handlePlay(s, m, []string{})
			}
		default:
			b.handleUnknown(s, m)
		}
	})
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error establishing discord connection: %v", err)
	}
	defer dg.Close()
	go StartEventServer(dg, b)
	fmt.Println("Ajzak turned on successfully!")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
	fmt.Println("Ajzak shutting down!")
}
