package bot

import (
	"sync"
	"time"
)

type Song struct {
	Title     string
	URL       string
	Duration  float64
	Uploader  string
	Thumbnail string
	Requester string
}

type GuildQueue struct {
	mu                sync.Mutex
	GuildID           string
	ChannelID         string
	Currently         *Song
	Songs             []*Song
	IsPlaying         bool
	idleTimer         *time.Timer
	LastTextChannelID string
}

type QueueManager struct {
	mu     sync.Mutex
	queues map[string]*GuildQueue
}

var GlobalQueueManager = &QueueManager{
	queues: make(map[string]*GuildQueue),
}

func (qm *QueueManager) Get(guildID string) *GuildQueue {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	if q, exists := qm.queues[guildID]; exists {
		return q
	}
	q := &GuildQueue{
		GuildID: guildID,
		Songs:   make([]*Song, 0),
	}
	qm.queues[guildID] = q
	return q
}

func (q *GuildQueue) Add(song *Song, textChannelID string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.ChannelID = textChannelID
	q.Songs = append(q.Songs, song)
}

func (q *GuildQueue) Next() *Song {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.Songs) == 0 {
		q.Currently = nil
		q.IsPlaying = false
		return nil
	}
	nextSong := q.Songs[0]
	q.Songs = q.Songs[1:]
	q.Currently = nextSong
	q.IsPlaying = true
	return nextSong
}

func (q *GuildQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.Songs = make([]*Song, 0)
	q.Currently = nil
	q.IsPlaying = false
}
