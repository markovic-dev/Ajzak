package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	data := i.MessageComponentData()
	if strings.HasPrefix(data.CustomID, "search_select_") {
		expectedUserID := strings.TrimPrefix(data.CustomID, "search_select_")
		if i.Member.User.ID != expectedUserID {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Search menu was not initiated by you!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		selectedURL := data.Values[0]
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "✅ Track selected!",
				Components: []discordgo.MessageComponent{},
			},
		})
		voiceState, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
		if err != nil || voiceState == nil {
			s.ChannelMessageSend(i.ChannelID, "❌ You must be in a voice channel!")
			return
		}
		meta, err := FetchMetadata(selectedURL)
		if err != nil {
			s.ChannelMessageSend(i.ChannelID, "❌ Failed to retrieve track metadata.")
			return
		}
		song := &Song{
			Title:     meta.Title,
			URL:       meta.WebpageURL,
			Duration:  meta.Duration,
			Uploader:  meta.Uploader,
			Thumbnail: meta.Thumbnail,
			Requester: i.Member.User.Mention(),
		}
		b.enqueueSong(s, i.GuildID, i.ChannelID, voiceState.ChannelID, song)
	}
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
	}, b)
	if !success {
		s.ChannelMessageSend(m.ChannelID, "Node.js service Error")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Joined the voice channel")
}

func (b *Bot) handlePlay(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "❌ You must provide a track URL or search query!")
		return
	}
	channelID := b.findUserVoiceChannel(s, m)
	if channelID == "" {
		s.ChannelMessageSend(m.ChannelID, "❌ You must be in a voice channel!")
		return
	}
	input := strings.Join(args, " ")
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		b.processPlayURL(s, m, input, channelID)
		return
	}
	loadingMsg, _ := s.ChannelMessageSend(m.ChannelID, "🔍 Fetching track metadata...")
	results, err := SearchTracks(input)
	if loadingMsg != nil {
		s.ChannelMessageDelete(m.ChannelID, loadingMsg.ID)
	}
	if err != nil || len(results) == 0 {
		s.ChannelMessageSend(m.ChannelID, "❌ Failed to retrieve track metadata.")
		return
	}
	var options []discordgo.SelectMenuOption
	for i, track := range results {
		title := track.Title
		if len(title) > 95 {
			title = title[:92] + "..."
		}
		options = append(options, discordgo.SelectMenuOption{
			Label:       fmt.Sprintf("%d. %s", i+1, title),
			Value:       track.URL,
			Description: fmt.Sprintf("%s | %s", track.Uploader, formatDuration(track.Duration)),
		})
	}
	selectMenu := discordgo.SelectMenu{
		CustomID:    "search_select_" + m.Author.ID,
		Placeholder: "Select a track from the list...",
		Options:     options,
	}
	s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("🔍 **Search results for:** `%s`\n*Select the correct one*", input),
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{selectMenu},
			},
		},
	})
}

func (b *Bot) handleSkip(s *discordgo.Session, m *discordgo.MessageCreate) {
	q := GlobalQueueManager.Get(m.GuildID)
	q.mu.Lock()
	q.IsReplay = false
	q.mu.Unlock()
	sendVoiceRequest("/stop", map[string]string{
		"guildId": m.GuildID,
	}, b)
	s.ChannelMessageSend(m.ChannelID, "⏭️ **Skipping track...**")
	PlayNextSong(s, m.GuildID, b)
}

func (b *Bot) handleQueue(s *discordgo.Session, m *discordgo.MessageCreate) {
	q := GlobalQueueManager.Get(m.GuildID)
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.Currently == nil && len(q.Songs) == 0 {
		s.ChannelMessageSend(m.ChannelID, "📜 The queue is empty.")
		return
	}
	var sb strings.Builder
	if q.Currently != nil {
		sb.WriteString(fmt.Sprintf("**Currently Playing:**\n▶️ [%s](%s) | `%s`\n\n",
			q.Currently.Title, q.Currently.URL, formatDuration(q.Currently.Duration)))
	}
	if len(q.Songs) > 0 {
		sb.WriteString("**Upcoming Tracks:**\n")
		for i, song := range q.Songs {
			if i >= 10 {
				sb.WriteString(fmt.Sprintf("\n*...and %d more tracks*", len(q.Songs)-10))
				break
			}
			sb.WriteString(fmt.Sprintf("`%d.` [%s](%s) | `%s`\n",
				i+1, song.Title, song.URL, formatDuration(song.Duration)))
		}
	}
	embed := &discordgo.MessageEmbed{
		Title:       "📜 Queue",
		Description: sb.String(),
		Color:       0x5865F2,
	}
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func (b *Bot) handleStop(s *discordgo.Session, m *discordgo.MessageCreate) {
	q := GlobalQueueManager.Get(m.GuildID)
	q.mu.Lock()
	q.IsReplay = false
	q.Songs = nil
	q.Currently = nil
	q.IsPlaying = false
	q.mu.Unlock()
	q.Clear()
	sendVoiceRequest("/stop", map[string]string{
		"guildId": m.GuildID,
	}, b)
	q.StartIdleTimer(s, m.GuildID, m.ChannelID, b)
	s.ChannelMessageSend(m.ChannelID, "⏹️ Playback stopped and queue cleared.")
}

func (b *Bot) handleLeave(s *discordgo.Session, m *discordgo.MessageCreate) {
	q := GlobalQueueManager.Get(m.GuildID)
	q.mu.Lock()
	q.IsReplay = false
	q.Songs = nil
	q.Currently = nil
	q.IsPlaying = false
	q.mu.Unlock()
	q.StopIdleTimer()
	success := sendVoiceRequest("/leave", map[string]string{
		"guildId": m.GuildID,
	}, b)
	if !success {
		s.ChannelMessageSend(m.ChannelID, "Not in a voice channel or a node service error.")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Left the voice channel.")
}

func (b *Bot) handleReplay(s *discordgo.Session, m *discordgo.MessageCreate) {
	q := GlobalQueueManager.Get(m.GuildID)
	q.mu.Lock()
	if !q.IsPlaying || q.Currently == nil {
		q.mu.Unlock()
		s.ChannelMessageSend(m.ChannelID, "❌ No song is currently streaming.")
		return
	}
	q.IsReplay = !q.IsReplay
	isReplayActive := q.IsReplay
	songTitle := q.Currently.Title
	q.mu.Unlock()
	if isReplayActive {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔂 **Replay On ** : **%s**", songTitle))
	} else {
		s.ChannelMessageSend(m.ChannelID, "▶️ **Replay Off**.")
	}
}
