package bot

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"text/tabwriter"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) handleUnknown(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "Unknown command!")
	s.ChannelMessageSend(m.ChannelID, "Available commands: `.provala`, `.stats`, `.dodaj <tekst>`")
}

func (b *Bot) handleProvala(s *discordgo.Session, m *discordgo.MessageCreate) {
	msg, err := b.Queries.GetLeastUsedProvale(context.Background())
	if err != nil {
		log.Printf("Unable to fetch responses from the database: %v", err)
		return
	}
	index := rand.IntN(len(msg))
	resp := msg[index]
	s.ChannelMessageSend(m.ChannelID, resp.Tekst)
	b.Queries.IncrementProvalaUsage(context.Background(), resp.ID)
}

func (b *Bot) handleStats(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := context.Background()
	stats, err := b.Queries.GetAllProvaleStats(ctx)
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

func (b *Bot) handleDodaj(s *discordgo.Session, m *discordgo.MessageCreate) {
	msg := strings.TrimSpace(strings.TrimPrefix(m.Content, ".dodaj "))
	if msg == "" {
		s.ChannelMessageSend(m.ChannelID, "Invalid input, example: `.dodaj <quoute>`")
		return
	}
	ctx := context.Background()
	err := b.Queries.InsertProvala(ctx, msg)
	if err != nil {
		log.Printf("Error adding qoute: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error adding qoute.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Added new qoute: \"%s\"", msg))
}
