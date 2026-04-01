package client_test

import (
	"testing"
	"time"

	"github.com/kqnade/vrcgo/shared"
	"github.com/kqnade/vrcli/internal/client"
)

func TestTrustRankFromTags(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want client.TrustRank
	}{
		{"no tags → Visitor", nil, client.TrustVisitor},
		{"basic → NewUser", []string{"system_trust_basic"}, client.TrustNewUser},
		{"known → User", []string{"system_trust_known"}, client.TrustUser},
		{"trusted → KnownUser", []string{"system_trust_trusted"}, client.TrustKnownUser},
		{"veteran → Trusted", []string{"system_trust_veteran"}, client.TrustTrusted},
		{"highest wins", []string{"system_trust_basic", "system_trust_veteran"}, client.TrustTrusted},
		{"unknown tag → Visitor", []string{"system_something_else"}, client.TrustVisitor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.TrustRankFromTags(tt.tags)
			if got != tt.want {
				t.Errorf("TrustRankFromTags(%v) = %v, want %v", tt.tags, got, tt.want)
			}
		})
	}
}

func TestFromLimitedUser_TrustFriendPriority(t *testing.T) {
	u := shared.LimitedUser{
		ID:          "usr_1",
		DisplayName: "Alice",
		Status:      "active",
		Tags:        []string{"system_trust_basic"},
		IsFriend:    true,
	}
	f := client.FromLimitedUser(u)
	if f.TrustRank != client.TrustFriend {
		t.Errorf("IsFriend=true should set TrustFriend, got %v", f.TrustRank)
	}
}

func TestFromLimitedUser_Fields(t *testing.T) {
	u := shared.LimitedUser{
		ID:                "usr_2",
		DisplayName:       "Bob",
		Status:            "joinMe",
		StatusDescription: "Playing now",
		Location:          "wrld_abc:12345",
		LastPlatform:      "standalonewindows",
		Tags:              []string{"system_trust_trusted"},
	}
	f := client.FromLimitedUser(u)

	if f.ID != u.ID {
		t.Errorf("ID = %q, want %q", f.ID, u.ID)
	}
	if f.DisplayName != u.DisplayName {
		t.Errorf("DisplayName = %q, want %q", f.DisplayName, u.DisplayName)
	}
	if f.Status != client.UserStatus(u.Status) {
		t.Errorf("Status = %q, want %q", f.Status, u.Status)
	}
	if f.StatusDescription != u.StatusDescription {
		t.Errorf("StatusDescription = %q, want %q", f.StatusDescription, u.StatusDescription)
	}
	if f.TrustRank != client.TrustKnownUser {
		t.Errorf("TrustRank = %v, want TrustKnownUser", f.TrustRank)
	}
	if f.Location != u.Location {
		t.Errorf("Location = %q, want %q", f.Location, u.Location)
	}
	if f.LastPlatform != u.LastPlatform {
		t.Errorf("LastPlatform = %q, want %q", f.LastPlatform, u.LastPlatform)
	}
}

func TestFromNotification_Fields(t *testing.T) {
	n := shared.Notification{
		ID:             "notif_1",
		Type:           "friendRequest",
		SenderUserID:   "usr_3",
		SenderUsername: "Charlie",
		Message:        "Hello!",
		Seen:           false,
		CreatedAt:      "2025-01-15T10:30:00.000Z",
	}
	notif := client.FromNotification(n)

	if notif.ID != n.ID {
		t.Errorf("ID = %q, want %q", notif.ID, n.ID)
	}
	if notif.Type != n.Type {
		t.Errorf("Type = %q, want %q", notif.Type, n.Type)
	}
	if !notif.IsFriendRequest {
		t.Error("IsFriendRequest should be true for type=friendRequest")
	}
	if notif.CreatedAt.IsZero() {
		t.Error("CreatedAt should be parsed from valid timestamp")
	}
	if notif.CreatedAtText == "" {
		t.Error("CreatedAtText should not be empty")
	}
}

func TestFromNotification_InvalidTimestamp(t *testing.T) {
	raw := "not-a-timestamp"
	n := shared.Notification{
		ID:        "notif_2",
		CreatedAt: raw,
	}
	notif := client.FromNotification(n)

	if !notif.CreatedAt.IsZero() {
		t.Error("CreatedAt should be zero for invalid timestamp")
	}
	// フォールバックとして元文字列を保持する
	if notif.CreatedAtText != raw {
		t.Errorf("CreatedAtText = %q, want %q (raw fallback)", notif.CreatedAtText, raw)
	}
}

func TestUserStatusEmoji(t *testing.T) {
	tests := []struct {
		status client.UserStatus
		emoji  string
	}{
		{client.StatusActive, "🟢"},
		{client.StatusJoinMe, "🔵"},
		{client.StatusAskMe, "🟡"},
		{client.StatusBusy, "🔴"},
		{client.StatusOffline, "⚫"},
	}
	for _, tt := range tests {
		if got := tt.status.Emoji(); got != tt.emoji {
			t.Errorf("(%q).Emoji() = %q, want %q", tt.status, got, tt.emoji)
		}
	}
}

func TestFromNotification_NotFriendRequest(t *testing.T) {
	n := shared.Notification{
		Type:      "invite",
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	notif := client.FromNotification(n)
	if notif.IsFriendRequest {
		t.Error("IsFriendRequest should be false for type=invite")
	}
}
