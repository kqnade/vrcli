# CLAUDE.md — vrchat-tui

VRChat の TUI ダッシュボード。Go + Bubbletea で実装する。

---

## プロジェクト概要

| 項目 | 内容 |
|------|------|
| 言語 | Go 1.25.5+ |
| TUI フレームワーク | [Bubbletea](https://github.com/charmbracelet/bubbletea) |
| スタイリング | [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| VRChat クライアント | [vrcgo](https://github.com/kqnade/vrcgo)（自作ライブラリ） |
| 認証情報管理 | `~/.config/vrchat-tui/session.json` + `~/.config/vrchat-tui/config.toml` |

---

## ディレクトリ構成

```
vrchat-tui/
├── cmd/vrchat/
│   └── main.go              # エントリポイント。cobra でサブコマンド管理
├── internal/
│   ├── ui/
│   │   ├── app.go           # ルート Model (Bubbletea)、ペイン切り替え
│   │   ├── friends.go       # フレンドリスト pane
│   │   ├── notifs.go        # 通知 pane (invite / friendreq)
│   │   ├── status.go        # ステータス変更 modal
│   │   ├── styles.go        # Lipgloss スタイル定義（ここだけで完結させる）
│   │   └── keys.go          # キーバインド定義 (bubbles/key)
│   ├── client/
│   │   ├── vrc.go           # vrcgo ラッパー + tea.Cmd ファクトリ
│   │   └── models.go        # UI 向けに整形したデータ型
│   └── config/
│       └── config.go        # 設定読み書き、認証情報管理
├── go.mod
├── go.sum
└── CLAUDE.md
```

---

## アーキテクチャ方針

### Bubbletea の使い方

- ルート Model は `internal/ui/app.go` の `AppModel`
- 各ペイン（friends / notifs / status）は独立した Model として実装し、`AppModel` が委譲する
- ペイン間のフォーカス切り替えは `tab` / `shift+tab`
- Modal（ステータス変更）はフォーカスを奪うオーバーレイとして実装
- **子モデルは vrc を持たない**。API 呼び出しの意図を意図メッセージ型で AppModel に伝え、AppModel が Cmd を発行する

### ポーリング

- `tea.Tick` で定期的に VRChat API を叩く
- 結果は `tea.Cmd` 経由でメッセージとして Model に流す
- デフォルトポーリング間隔：フレンド 30s、通知 15s

### エラーハンドリング

- API エラーは `ErrMsg` 型にラップして Model に流す
- 認証エラー（401）は `errors.As` で `shared.APIError` を取り出し `StatusCode == 401` で判定
- TUI 上ではステータスバーにエラーを一時表示（3秒後に消す）

---

## キーバインド

| キー | 動作 |
|------|------|
| `tab` / `shift+tab` | ペイン切り替え |
| `j` / `k` または `↑` / `↓` | リスト移動 |
| `enter` | 選択・決定 |
| `s` | ステータス変更 modal を開く |
| `r` | 手動リフレッシュ |
| `q` / `ctrl+c` | 終了 |
| `?` | ヘルプ表示 |

---

## スタイリング

- カラーパレットは `internal/ui/styles.go` に集約する
- ダークターミナル前提。ライトテーマ考慮は不要
- Trust Rank は色で区別する（Visitor=灰、New User=青、User=緑、Known User=橙、Trusted=紫、Friend=黄）
- ステータスインジケーター：🟢 Online / 🟡 Ask Me / 🔴 Busy / ⚫ Offline

---

## 認証フロー

1. `~/.config/vrchat-tui/session.json` に保存済みセッションがあれば使う
2. なければ `vrchat-tui auth` サブコマンドを案内
3. auth: 対話形式でユーザー名/パスワード入力 → 2FA 必要なら TOTP 追加入力 → セッション保存

---

## コーディング規約

- `gofmt` / `goimports` を必ず通す
- エクスポートする型・関数には godoc コメントをつける
- `internal/` 以下はパッケージをまたいだ直接参照を避ける（client → ui の依存は禁止）
- エラーは `fmt.Errorf("context: %w", err)` でラップして伝播させる
- ゴルーチンのリークに注意。`context.Context` でキャンセルを伝播させる

---

## ビルド・実行

```bash
go build ./cmd/vrchat
./vrchat          # TUI 起動
./vrchat auth     # 認証情報セットアップ
./vrchat version  # バージョン表示
```
