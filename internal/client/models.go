// Package client は vrcgo ラッパーと UI 向けデータ型を提供する。
package client

import (
	"time"

	"github.com/kqnade/vrcgo/shared"
)

// TrustRank は VRChat のトラストランクを表す。
type TrustRank int

const (
	// TrustVisitor は未認証ユーザー。
	TrustVisitor TrustRank = iota
	// TrustNewUser は新規ユーザー。
	TrustNewUser
	// TrustUser は通常ユーザー。
	TrustUser
	// TrustKnownUser は既知ユーザー。
	TrustKnownUser
	// TrustTrusted は信頼されたユーザー。
	TrustTrusted
	// TrustFriend はフレンド（最優先）。
	TrustFriend
)

// UserStatus はユーザーのオンラインステータスを表す。
type UserStatus string

const (
	// StatusActive はオンライン状態。
	StatusActive UserStatus = "active"
	// StatusJoinMe は Join Me 状態。
	StatusJoinMe UserStatus = "joinMe"
	// StatusAskMe は Ask Me 状態。
	StatusAskMe UserStatus = "askMe"
	// StatusBusy は Busy 状態。
	StatusBusy UserStatus = "busy"
	// StatusOffline はオフライン状態。
	StatusOffline UserStatus = "offline"
)

// Emoji はステータスに対応する絵文字を返す。
func (s UserStatus) Emoji() string {
	switch s {
	case StatusActive, StatusJoinMe:
		return "🟢"
	case StatusAskMe:
		return "🟡"
	case StatusBusy:
		return "🔴"
	default:
		return "⚫"
	}
}

// Friend は UI 表示用のフレンドデータを保持する。
type Friend struct {
	ID                string
	DisplayName       string
	Status            UserStatus
	StatusDescription string
	Location          string
	LastPlatform      string
	TrustRank         TrustRank
}

// Notif は UI 表示用の通知データを保持する。
type Notif struct {
	ID              string
	Type            string
	SenderUserID    string
	SenderUsername  string
	Message         string
	Seen            bool
	CreatedAt       time.Time
	CreatedAtText   string // 表示用整形済み文字列。パース失敗時は元の文字列
	IsFriendRequest bool
}

// TrustRankFromTags は vrcgo のタグスライスから TrustRank を算出する。
// 複数タグがある場合は最も高いランクを返す。
func TrustRankFromTags(tags []string) TrustRank {
	rank := TrustVisitor
	for _, tag := range tags {
		switch tag {
		case "system_trust_veteran":
			if TrustTrusted > rank {
				rank = TrustTrusted
			}
		case "system_trust_trusted":
			if TrustKnownUser > rank {
				rank = TrustKnownUser
			}
		case "system_trust_known":
			if TrustUser > rank {
				rank = TrustUser
			}
		case "system_trust_basic":
			if TrustNewUser > rank {
				rank = TrustNewUser
			}
		}
	}
	return rank
}

// FromLimitedUser は shared.LimitedUser を Friend に変換する。
// IsFriend が true の場合は TrustFriend を優先する。
func FromLimitedUser(u shared.LimitedUser) Friend {
	rank := TrustRankFromTags(u.Tags)
	if u.IsFriend {
		rank = TrustFriend
	}
	return Friend{
		ID:                u.ID,
		DisplayName:       u.DisplayName,
		Status:            UserStatus(u.Status),
		StatusDescription: u.StatusDescription,
		Location:          u.Location,
		LastPlatform:      u.LastPlatform,
		TrustRank:         rank,
	}
}

// FromNotification は shared.Notification を Notif に変換する。
// CreatedAt のパースに失敗した場合は CreatedAtText に元文字列を格納する。
func FromNotification(n shared.Notification) Notif {
	notif := Notif{
		ID:              n.ID,
		Type:            n.Type,
		SenderUserID:    n.SenderUserID,
		SenderUsername:  n.SenderUsername,
		Message:         n.Message,
		Seen:            n.Seen,
		IsFriendRequest: n.Type == "friendRequest",
	}

	t, err := time.Parse(time.RFC3339, n.CreatedAt)
	if err != nil {
		// パース失敗時は元文字列をそのまま保持
		notif.CreatedAtText = n.CreatedAt
	} else {
		notif.CreatedAt = t
		notif.CreatedAtText = t.Format("2006-01-02 15:04")
	}
	return notif
}
