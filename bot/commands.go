package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"

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

//MUSIC FUNCIONALITY

func (b *Bot) findUserVoiceChannel(s *discordgo.Session, m *discordgo.MessageCreate) string {
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		return ""
	}
	for _, vs := range guild.VoiceStates {
		if vs.UserID == m.Author.ID {
			return vs.ChannelID
		}
	}
	return ""
}

func sendVoiceRequest(endpoint string, payload map[string]string) bool {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false
	}

	resp, err := http.Post("http://localhost:3000"+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()

	return true
}

type SongMetadata struct {
	Title      string  `json:"title"`
	Uploader   string  `json:"uploader"`
	Duration   float64 `json:"duration"`
	Thumbnail  string  `json:"thumbnail"`
	WebpageURL string  `json:"webpage_url"`
}

func FetchMetadata(url string) (*SongMetadata, error) {
	cmd := exec.Command("yt-dlp", "-j", "--no-playlist", url)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	var meta SongMetadata
	if err := json.Unmarshal(out.Bytes(), &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func formatDuration(seconds float64) string {
	d := time.Duration(seconds) * time.Second
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d : %02d", m, s)
}

func (b *Bot) handleJoin(s *discordgo.Session, m *discordgo.MessageCreate) {
	channelID := b.findUserVoiceChannel(s, m)
	if channelID == "" {
		s.ChannelMessageSend(m.ChannelID, "Error user must be in a voice channel.")
		return
	}
	success := sendVoiceRequest("/join", map[string]string{
		"guildId":   m.GuildID,
		"channelId": channelID,
	})
	if !success {
		s.ChannelMessageSend(m.ChannelID, "Node.js service Error")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Joined the voice channel")
}

func (b *Bot) handlePlay(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command MUST include a song name or URL")
		return
	}
	url := args[0]
	channelID := b.findUserVoiceChannel(s, m)
	if channelID == "" {
		s.ChannelMessageSend(m.ChannelID, "Error user must be in a voice channel.")
		return
	}
	loadingMsg, _ := s.ChannelMessageSend(m.ChannelID, "🔍 Fetching song info")
	meta, err := FetchMetadata(url)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching song info.")
		return
	}
	sendVoiceRequest("/join", map[string]string{
		"guildId":   m.GuildID,
		"channelId": channelID,
	})
	success := sendVoiceRequest("/play", map[string]string{
		"guildId": m.GuildID,
		"url":     meta.WebpageURL,
	})
	if loadingMsg != nil {
		s.ChannelMessageDelete(m.ChannelID, loadingMsg.ID)
	}
	if !success {
		s.ChannelMessageSend(m.ChannelID, "Streaming error.")
		return
	}
	guild, _ := s.State.Guild(m.GuildID)
	guildName := "serveru"
	if guild != nil {
		guildName = guild.Name
	}
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("▶️ Playing in %s", guildName),
		Description: fmt.Sprintf("**[%s](%s)**\nby **%s**",
			meta.Title, meta.WebpageURL, meta.Uploader),
		Color: 0x2B2D31,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: meta.Thumbnail,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Played by",
				Value:  m.Author.Mention(),
				Inline: true,
			},
			{
				Name:   "Track Duration",
				Value:  fmt.Sprintf("` %s `", formatDuration(meta.Duration)),
				Inline: true,
			},
			{
				Name:   "Songs in Queue",
				Value:  "` 0 `",
				Inline: true,
			},
		},
	}
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
func (b *Bot) handleLeave(s *discordgo.Session, m *discordgo.MessageCreate) {
	success := sendVoiceRequest("/leave", map[string]string{
		"guildId": m.GuildID,
	})
	if !success {
		s.ChannelMessageSend(m.ChannelID, "Not in a voice channel or a node service error.")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Left the voice channel.")
}
