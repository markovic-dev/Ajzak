package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
)

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

func sendVoiceRequest(endpoint string, payload map[string]string, b *Bot) bool {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	resp, err := http.Post(b.NodeServiceURL+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()
	return true
}

func (b *Bot) processPlayURL(s *discordgo.Session, m *discordgo.MessageCreate, url string, voiceChannelID string) {
	loadingMsg, _ := s.ChannelMessageSend(m.ChannelID, "🔍 Fetching track metadata...")
	meta, err := FetchMetadata(url)
	if loadingMsg != nil {
		s.ChannelMessageDelete(m.ChannelID, loadingMsg.ID)
	}
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Failed to retrieve track metadata.")
		return
	}
	song := &Song{
		Title:     meta.Title,
		URL:       meta.WebpageURL,
		Duration:  meta.Duration,
		Uploader:  meta.Uploader,
		Thumbnail: meta.Thumbnail,
		Requester: m.Author.Mention(),
	}
	b.enqueueSong(s, m.GuildID, m.ChannelID, voiceChannelID, song)
}

func (b *Bot) enqueueSong(s *discordgo.Session, guildID, textChannelID, voiceChannelID string, song *Song) {
	q := GlobalQueueManager.Get(guildID)
	q.StopIdleTimer()
	q.mu.Lock()
	q.LastTextChannelID = textChannelID
	isPlaying := q.IsPlaying
	q.mu.Unlock()
	if isPlaying {
		q.mu.Lock()
		q.Songs = append(q.Songs, song)
		queuePos := len(q.Songs)
		q.mu.Unlock()
		embed := &discordgo.MessageEmbed{
			Title:       "📝 Added to Queue",
			Description: fmt.Sprintf("**[%s](%s)**", song.Title, song.URL),
			Color:       0x00FF00,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: song.Thumbnail,
			},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Queue Position",
					Value:  fmt.Sprintf("` #%d `", queuePos),
					Inline: true,
				},
				{
					Name:   "Track Duration",
					Value:  fmt.Sprintf("` %s `", formatDuration(song.Duration)),
					Inline: true,
				},
			},
		}
		s.ChannelMessageSendEmbed(textChannelID, embed)
		return
	}
	q.mu.Lock()
	q.Songs = append(q.Songs, song)
	q.mu.Unlock()
	sendVoiceRequest("/join", map[string]string{
		"guildId":   guildID,
		"channelId": voiceChannelID,
	}, b)
	PlayNextSong(s, guildID, b)
}

func formatDuration(seconds float64) string {
	d := time.Duration(seconds) * time.Second
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d : %02d", m, s)
}

func (q *GuildQueue) StartIdleTimer(s *discordgo.Session, guildID string, textChannelID string, b *Bot) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.idleTimer != nil {
		q.idleTimer.Stop()
	}
	q.idleTimer = time.AfterFunc(1*time.Minute, func() {
		q.mu.Lock()
		if q.IsPlaying || len(q.Songs) > 0 {
			q.mu.Unlock()
			return
		}
		q.mu.Unlock()
		sendVoiceRequest("/leave", map[string]string{
			"guildId": guildID,
		}, b)
		if textChannelID != "" {
			s.ChannelMessageSend(textChannelID, "⏱️ Left the voice channel due to inactivity.")
		}
	})
}
func (q *GuildQueue) StopIdleTimer() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.idleTimer != nil {
		q.idleTimer.Stop()
		q.idleTimer = nil
	}
}
