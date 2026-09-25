package bot

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/markovic-dev/ajzak/internal/database"
	_ "modernc.org/sqlite"
)

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
		if m.Content == ".provala" {
			msg, err := queries.GetLeastUsedProvale(context.Background())
			if err != nil {
				log.Printf("Unable to fetch responses from the database: %v", err)
				return
			}
			index := rand.IntN(len(msg))
			resp := msg[index]
			s.ChannelMessageSend(m.ChannelID, resp.Tekst)
			queries.IncrementProvalaUsage(context.Background(), resp.ID)
		}

		if m.Content == ".stats" {
			ctx := context.Background()
			stats, err := queries.GetAllProvaleStats(ctx)
			if err != nil {
				log.Printf("Error fetching statistics: %v", err)
				s.ChannelMessageSend(m.ChannelID, "Unable to fetch stats.")
				return
			}

			var buf bytes.Buffer
			w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

			fmt.Fprintln(w, "ID\tProvala\tPuta")
			fmt.Fprintln(w, "--\t-------\t---------")

			for _, p := range stats {
				msg := p.Tekst
				if len(msg) > 30 {
					msg = msg[:27] + "..."
				}
				fmt.Fprintf(w, "%d\t%s\t%dx\n", p.ID, msg, p.UsageCount)
			}
			w.Flush()

			resp := "```\n" + buf.String() + "```"
			s.ChannelMessageSend(m.ChannelID, resp)
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
	fmt.Println("Ajzak shutting down!")
	<-sc
}
