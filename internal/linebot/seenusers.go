package linebot

import (
	"sort"
	"sync"
	"time"
)

// SeenUser is a LINE user that has recently messaged the bot. It powers the
// receiver picker in the web UI so users don't have to copy IDs by hand.
type SeenUser struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	LastSeen time.Time `json:"lastSeen"`
}

const maxSeenUsers = 50

var (
	seenMu    sync.Mutex
	seenUsers = make(map[string]SeenUser)
)

// recordSeenUser upserts a user, updating its last-seen time and (if provided)
// its display name. The store is capped, dropping the least-recently-seen user.
func recordSeenUser(id, name string) {
	seenMu.Lock()
	defer seenMu.Unlock()

	u := seenUsers[id]
	u.ID = id
	if name != "" {
		u.Name = name
	}
	u.LastSeen = time.Now()
	seenUsers[id] = u

	if len(seenUsers) > maxSeenUsers {
		var oldestID string
		var oldest time.Time
		first := true
		for k, v := range seenUsers {
			if first || v.LastSeen.Before(oldest) {
				oldest, oldestID, first = v.LastSeen, k, false
			}
		}
		delete(seenUsers, oldestID)
	}
}

// seenUserNeedsName reports whether we still lack a display name for the user.
func seenUserNeedsName(id string) bool {
	seenMu.Lock()
	defer seenMu.Unlock()
	u, ok := seenUsers[id]
	return !ok || u.Name == ""
}

// SeenUsers returns recently-seen users, most recent first.
func SeenUsers() []SeenUser {
	seenMu.Lock()
	defer seenMu.Unlock()

	out := make([]SeenUser, 0, len(seenUsers))
	for _, u := range seenUsers {
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastSeen.After(out[j].LastSeen) })
	return out
}
