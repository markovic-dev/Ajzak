package bot

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

type SongEndedEvent struct {
	GuildID string `json:"guildId"`
}

func StartEventServer(dg *discordgo.Session, b *Bot) {

	http.HandleFunc("/events/song-ended", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}
		var evt SongEndedEvent
		if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		go PlayNextSong(dg, evt.GuildID, b)

		w.WriteHeader(http.StatusOK)
	})
	fmt.Println("Go Event Listener on http://localhost:8080")
	go http.ListenAndServe(":8080", nil)
}

func PlayNextSong(s *discordgo.Session, guildID string, b *Bot) {
	q := GlobalQueueManager.Get(guildID)
	q.mu.Lock()
	if q.IsReplay && q.Currently != nil {
		song := q.Currently
		textChannelID := q.LastTextChannelID
		q.mu.Unlock()
		success := sendVoiceRequest("/play", map[string]string{
			"guildId": guildID,
			"url":     song.URL,
		}, b)
		if !success {
			if textChannelID != "" {
				s.ChannelMessageSend(textChannelID, "❌ Error replaying song.")
			}
			q.mu.Lock()
			q.IsPlaying = false
			q.IsReplay = false
			q.mu.Unlock()
			return
		}
		return
	}
	if len(q.Songs) == 0 {
		q.IsPlaying = false
		q.Currently = nil
		lastChannel := q.LastTextChannelID
		q.mu.Unlock()
		q.StartIdleTimer(s, guildID, lastChannel, b)
		return
	}
	song := q.Songs[0]
	q.Songs = q.Songs[1:]
	q.Currently = song
	q.IsPlaying = true
	textChannelID := q.LastTextChannelID
	q.mu.Unlock()
	success := sendVoiceRequest("/play", map[string]string{
		"guildId": guildID,
		"url":     song.URL,
	}, b)

	if !success {
		if textChannelID != "" {
			s.ChannelMessageSend(textChannelID, "❌ Error streaming music.")
		}
		q.mu.Lock()
		q.IsPlaying = false
		q.mu.Unlock()
		return
	}
	channelName := "voice"
	if g, err := s.State.Guild(guildID); err == nil {
		for _, vs := range g.VoiceStates {
			if vs.UserID == s.State.User.ID {
				if ch, err := s.Channel(vs.ChannelID); err == nil {
					channelName = ch.Name
				}
				break
			}
		}
	}
	if textChannelID != "" {
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("▶️ Now Playing in %s", channelName),
			Description: fmt.Sprintf("**[%s](%s)**\nby **%s**", song.Title, song.URL, song.Uploader),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: song.Thumbnail,
			},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Played by",
					Value:  song.Requester,
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
	}
}
