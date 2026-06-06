package linebot

import (
	"fmt"
	"testing"
	"time"
)

func resetSeenUsers() {
	seenMu.Lock()
	defer seenMu.Unlock()
	seenUsers = make(map[string]SeenUser)
}

func TestRecordSeenUserUpsert(t *testing.T) {
	resetSeenUsers()

	recordSeenUser("U1", "")
	if !seenUserNeedsName("U1") {
		t.Error("user recorded without a name should still need a name")
	}

	recordSeenUser("U1", "Alice")
	if seenUserNeedsName("U1") {
		t.Error("user should no longer need a name after recording one")
	}

	// An empty name must not wipe an existing name.
	recordSeenUser("U1", "")
	if seenUserNeedsName("U1") {
		t.Error("empty name should not clear an existing name")
	}

	users := SeenUsers()
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("name = %q, want Alice", users[0].Name)
	}
}

func TestSeenUserNeedsNameUnknown(t *testing.T) {
	resetSeenUsers()
	if !seenUserNeedsName("nope") {
		t.Error("unknown user should need a name")
	}
}

func TestRecordSeenUserEvictsOldest(t *testing.T) {
	resetSeenUsers()

	// Seed the store to its cap with controlled, strictly-increasing last-seen
	// times so "oldest" is unambiguous (real time.Now() can tie on coarse clocks).
	base := time.Now().Add(-time.Hour)
	seenMu.Lock()
	for i := range maxSeenUsers {
		id := fmt.Sprintf("U%03d", i)
		seenUsers[id] = SeenUser{ID: id, LastSeen: base.Add(time.Duration(i) * time.Minute)}
	}
	seenMu.Unlock()

	// Recording a new user pushes past the cap and must evict the oldest (U000).
	recordSeenUser("Unew", "")

	seenMu.Lock()
	n := len(seenUsers)
	_, oldestStillThere := seenUsers["U000"]
	_, newPresent := seenUsers["Unew"]
	seenMu.Unlock()

	if n != maxSeenUsers {
		t.Errorf("store size = %d, want %d", n, maxSeenUsers)
	}
	if oldestStillThere {
		t.Error("the oldest user (U000) should have been evicted")
	}
	if !newPresent {
		t.Error("the newly recorded user should be present")
	}
}

func TestSeenUsersOrderedMostRecentFirst(t *testing.T) {
	resetSeenUsers()

	// Seed distinct last-seen times directly so ordering is unambiguous.
	now := time.Now()
	seenMu.Lock()
	seenUsers["U3"] = SeenUser{ID: "U3", LastSeen: now.Add(-3 * time.Minute)}
	seenUsers["U1"] = SeenUser{ID: "U1", LastSeen: now}
	seenUsers["U2"] = SeenUser{ID: "U2", LastSeen: now.Add(-1 * time.Minute)}
	seenMu.Unlock()

	users := SeenUsers()
	want := []string{"U1", "U2", "U3"}
	if len(users) != len(want) {
		t.Fatalf("expected %d users, got %d", len(want), len(users))
	}
	for i, id := range want {
		if users[i].ID != id {
			t.Errorf("position %d = %q, want %q", i, users[i].ID, id)
		}
	}
}
