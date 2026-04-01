package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kqnade/vrcgo/shared"
	"github.com/kqnade/vrcgo/vrcapi"
)

// ErrTwoFactorRequired は 2FA コードが必要なときに返されるエラー。
var ErrTwoFactorRequired = errors.New("two-factor authentication required")

// tea.Msg 型群

// FriendsMsg はフレンドリスト取得成功時のメッセージ。
type FriendsMsg struct{ Friends []Friend }

// NotifsMsg は通知リスト取得成功時のメッセージ。
type NotifsMsg struct{ Notifs []Notif }

// ErrMsg はエラー発生時のメッセージ。
type ErrMsg struct {
	Err    error
	IsAuth bool // true なら 401 認証エラー
}

// AuthOKMsg は認証確認成功時のメッセージ。
type AuthOKMsg struct {
	UserID            string
	DisplayName       string
	Status            UserStatus
	StatusDescription string
}

// StatusUpdatedMsg はステータス更新成功時のメッセージ。
type StatusUpdatedMsg struct {
	Status     UserStatus
	StatusDesc string
}

// FriendRequestAcceptedMsg はフレンドリクエスト承認成功時のメッセージ。
type FriendRequestAcceptedMsg struct{ NotifID string }

// FriendRequestRejectedMsg はフレンドリクエスト拒否成功時のメッセージ。
type FriendRequestRejectedMsg struct{ NotifID string }

// VRCClient は vrcgo クライアントのラッパー。
type VRCClient struct {
	api    *vrcapi.Client
	userID string
}

const requestTimeout = 30 * time.Second

// New は新しい VRCClient を作成する。
func New(userAgent string) (*VRCClient, error) {
	api, err := vrcapi.NewClient(shared.WithUserAgent(userAgent))
	if err != nil {
		return nil, fmt.Errorf("client: create vrcapi client: %w", err)
	}
	return &VRCClient{api: api}, nil
}

// LoadSession はセッション Cookie ファイルを読み込む。
func (c *VRCClient) LoadSession(path string) error {
	if err := c.api.LoadCookies(path); err != nil {
		return fmt.Errorf("client: load session: %w", err)
	}
	return nil
}

// SaveSession はセッション Cookie をファイルに保存する。
// 必要に応じてディレクトリを作成する。
func (c *VRCClient) SaveSession(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("client: create session dir: %w", err)
	}
	if err := c.api.SaveCookies(path); err != nil {
		return fmt.Errorf("client: save session: %w", err)
	}
	return nil
}

// CheckSession はセッションの有効性を確認する tea.Cmd。
// 成功時は AuthOKMsg を返す。
func (c *VRCClient) CheckSession() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		user, err := c.api.GetCurrentUser(ctx)
		if err != nil {
			return ErrMsg{Err: err, IsAuth: isAuthError(err)}
		}
		c.userID = user.ID
		return AuthOKMsg{
			UserID:            user.ID,
			DisplayName:       user.DisplayName,
			Status:            UserStatus(user.Status),
			StatusDescription: user.StatusDescription,
		}
	}
}

// AuthenticateSync は認証を同期的に実行する（auth コマンド用）。
// 2FA が必要な場合は ErrTwoFactorRequired を返す。
func (c *VRCClient) AuthenticateSync(username, password, totp string) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	err := c.api.Authenticate(ctx, shared.AuthConfig{
		Username: username,
		Password: password,
		TOTPCode: totp,
	})
	if err != nil {
		if isTwoFactorError(err) {
			return ErrTwoFactorRequired
		}
		return fmt.Errorf("client: authenticate: %w", err)
	}
	return nil
}

// Authenticate は認証を行う tea.Cmd（TUI 用）。
func (c *VRCClient) Authenticate(username, password, totp string) tea.Cmd {
	return func() tea.Msg {
		if err := c.AuthenticateSync(username, password, totp); err != nil {
			return ErrMsg{Err: err, IsAuth: isAuthError(err)}
		}
		return c.CheckSession()()
	}
}

// FetchFriends はオンラインフレンドリストを取得する tea.Cmd。
func (c *VRCClient) FetchFriends() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		users, err := c.api.GetFriends(ctx, shared.GetFriendsOptions{
			Offline: false,
			N:       100,
		})
		if err != nil {
			return ErrMsg{Err: err, IsAuth: isAuthError(err)}
		}

		friends := make([]Friend, len(users))
		for i, u := range users {
			friends[i] = FromLimitedUser(u)
		}
		return FriendsMsg{Friends: friends}
	}
}

// FetchNotifs は通知リストを取得する tea.Cmd。
func (c *VRCClient) FetchNotifs() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		notifs, err := c.api.GetNotifications(ctx, shared.GetNotificationsOptions{
			N: 100,
		})
		if err != nil {
			return ErrMsg{Err: err, IsAuth: isAuthError(err)}
		}

		result := make([]Notif, len(notifs))
		for i, n := range notifs {
			result[i] = FromNotification(n)
		}
		return NotifsMsg{Notifs: result}
	}
}

// AcceptFriendRequest はフレンドリクエストを承認する tea.Cmd。
func (c *VRCClient) AcceptFriendRequest(notifID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		if _, err := c.api.AcceptFriendRequest(ctx, notifID); err != nil {
			return ErrMsg{Err: err, IsAuth: isAuthError(err)}
		}
		return FriendRequestAcceptedMsg{NotifID: notifID}
	}
}

// RejectFriendRequest はフレンドリクエストを拒否する tea.Cmd。
func (c *VRCClient) RejectFriendRequest(notifID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		if _, err := c.api.RejectFriendRequest(ctx, notifID); err != nil {
			return ErrMsg{Err: err, IsAuth: isAuthError(err)}
		}
		return FriendRequestRejectedMsg{NotifID: notifID}
	}
}

// UpdateStatus はユーザーステータスを更新する tea.Cmd。
// CheckSession または Authenticate を先に呼び出して userID を設定しておく必要がある。
func (c *VRCClient) UpdateStatus(status, desc string) tea.Cmd {
	return func() tea.Msg {
		if c.userID == "" {
			return ErrMsg{Err: fmt.Errorf("client: userID not set; call CheckSession first")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		s := status
		d := desc
		if _, err := c.api.UpdateUser(ctx, c.userID, shared.UpdateUserRequest{
			Status:            &s,
			StatusDescription: &d,
		}); err != nil {
			return ErrMsg{Err: err, IsAuth: isAuthError(err)}
		}
		return StatusUpdatedMsg{
			Status:     UserStatus(status),
			StatusDesc: desc,
		}
	}
}

// isAuthError は err が 401 認証エラーかどうかを判定する。
// vrcgo は fmt.Errorf("%w") でラップするため errors.As を使う。
func isAuthError(err error) bool {
	var apiErr *shared.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 401
}

// isTwoFactorError は err が 2FA 要求エラーかどうかを判定する。
func isTwoFactorError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "two-factor authentication required")
}
