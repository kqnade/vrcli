// Package main はエントリポイント。cobra で TUI / auth / version サブコマンドを提供する。
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/kqnade/vrcli/internal/client"
	"github.com/kqnade/vrcli/internal/config"
	"github.com/kqnade/vrcli/internal/ui"
)

const (
	appVersion = "0.1.0"
	userAgent  = "vrchat-tui/" + appVersion + " github.com/kqnade/vrcli"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:          "vrchat",
		Short:        "VRChat TUI ダッシュボード",
		SilenceUsage: true,
		RunE:         runTUI,
	}
	root.AddCommand(authCmd(), versionCmd())
	return root
}

// runTUI はセッションを読み込んで TUI を起動する。
func runTUI(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("設定の読み込みに失敗しました: %w", err)
	}

	vrc, err := client.New(userAgent)
	if err != nil {
		return fmt.Errorf("クライアントの初期化に失敗しました: %w", err)
	}

	sessionPath, err := config.SessionPath()
	if err != nil {
		return err
	}

	if err := vrc.LoadSession(sessionPath); err != nil {
		fmt.Fprintln(os.Stderr, "セッションが見つかりません。先に 'vrchat auth' を実行してください。")
		return err
	}

	model := ui.New(vrc, cfg)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI の実行に失敗しました: %w", err)
	}
	return nil
}

// authCmd は対話式認証サブコマンド。
func authCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "VRChat アカウントで認証してセッションを保存する",
		RunE:  runAuth,
	}
}

func runAuth(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("ユーザー名またはメールアドレス: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("入力エラー: %w", err)
	}
	username = strings.TrimSpace(username)

	fmt.Print("パスワード: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("パスワード入力エラー: %w", err)
	}
	password := string(passwordBytes)

	vrc, err := client.New(userAgent)
	if err != nil {
		return fmt.Errorf("クライアントの初期化に失敗しました: %w", err)
	}

	// 最初は TOTP なしで試みる
	authErr := vrc.AuthenticateSync(username, password, "")
	if authErr != nil {
		if authErr != client.ErrTwoFactorRequired {
			return fmt.Errorf("認証に失敗しました: %w", authErr)
		}

		// 2FA が必要
		fmt.Print("2FA コード (TOTP): ")
		totpBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("2FA コード入力エラー: %w", err)
		}
		totp := strings.TrimSpace(string(totpBytes))

		if authErr = vrc.AuthenticateSync(username, password, totp); authErr != nil {
			return fmt.Errorf("2FA 認証に失敗しました: %w", authErr)
		}
	}

	sessionPath, err := config.SessionPath()
	if err != nil {
		return err
	}

	if err := vrc.SaveSession(sessionPath); err != nil {
		return fmt.Errorf("セッションの保存に失敗しました: %w", err)
	}

	fmt.Println("認証成功！セッションを保存しました。")
	return nil
}

// versionCmd はバージョン表示サブコマンド。
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "バージョンを表示する",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("vrchat-tui version " + appVersion)
		},
	}
}
