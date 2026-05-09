# Skittles

X68000 用ファイラー [mint.x](https://opencode.ai) の設計思想を継承した、
モダンな 2画面 TUI ファイルマネージャ。

```
┌─ Skittles v0.1.0 ─────────── 2026/05/09 12:34:56 ─┐
├─ Left Pane ────────┬─ Right Pane ───────┐
│  📁 drwxr-xr-x Documents  │  📁 drwxr-xr-x Downloads  │
│  📄 -rw-r--r-- main.go    │  📄 -rw-r--r-- data.json  │
├────────────────────┴─────────────────────┤
│ Console                                  │
│ > ls -la                                 │
├──────────────────────────────────────────┤
│ ↑↓:nav | Enter:open | Tab:focus(3-pane) │
└──────────────────────────────────────────┘
```

## 特徴

- **2画面 + コンソール の3ペイン**: Tab でフォーカス切替、アクティブペインが source、反対が destination
- **ファイル操作**: コピー/移動/削除/マーク/ソート
- **パーミッション表示**: 各行に `drwxr-xr-x` 形式
- **ビルトインビューア**: テキスト・Markdown・ソースコードのシンタックスハイライト対応（スクロール可能）。**実行ビットなしファイルは自動でビューア表示**
- **拡張子→アクション**: `.go` → ビューア、`.zip` → unzip、`.mdx` → MP4M.app 等、`config.yaml` でカスタマイズ可能
- **コンソールペイン**: シェルコマンドの実行と**リアルタイム出力表示**、コマンド履歴
- **クロスプラットフォーム**: macOS / Linux / Windows（シングルバイナリ）
- **日本語入力対応**: macOS 起動時に自動で英数入力に切替
- **トップバー**: アプリ名・バージョン・リアルタイム時計表示
- **画像プレビュー**: `p` キーで外部アプリ（プレビュー.app等）で開く

## インストール

### Homebrew（準備中）

```bash
brew install ktam72/tap/skittles
```

### Go

```bash
go install github.com/ktam72/skittles@latest
```

### 手動ビルド

```bash
git clone git@github.com:ktam72/skittles.git
cd skittles
go build -o skittles .
./skittles
```

## 使い方

```bash
# カレントディレクトリで起動
./skittles

# 左右別のディレクトリで起動
./skittles /usr/local /opt/homebrew

# 同じディレクトリで起動
./skittles ~/Documents
```

### キーバインド

| キー | 機能 |
|------|------|
| `↑↓ / kj` | カーソル移動 |
| `→ / l` | ブラウズモードでは無効 |
| `Enter` | ディレクトリ進入 / ファイルを開く / 実行ファイルはコンソールにパス入力 |
| `← / h` | 親ディレクトリへ |
| `Tab` | フォーカス切替（Left→Right→Console→Left） |
| `Space / b` | ページ送り / 戻し |
| `Backspace` | マーク |
| `a` | 全マーク |
| `p` | 外部アプリでプレビュー（画像→プレビュー.app 等） |
| `c` | 反対ペインへコピー |
| `m` | 反対ペインへ移動 |
| `d` | 削除 |
| `sr` | ソート切替 |
| `!` | コンソールへジャンプ |
| `E` | エディタで開く |
| `ESC` | 終了（2回押し） |

### ビューア操作

| キー | 機能 |
|------|------|
| `↑↓` | 1行スクロール |
| `PgUp / PgDown` | 1ページスクロール |
| `ESC` | 閉じる |

## 設定

`config.yaml` で拡張子ごとのアクションをカスタマイズできます。

```yaml
actions:
  - match: ".md"
    viewer: true
  - match: ".zip"
    command: "unzip -o $P"
  - match: ".mdx"
    command: "open -a MP4M.app $P"
```

変数: `$P` = フルパス、`$F` = ファイル名、`$D` = ディレクトリ、`$EDITOR` = エディタ

## 開発

```bash
go build -o skittles .              # ビルド
go vet ./...                         # 静的解析
golangci-lint run ./...              # Lint（0 issues）

# クロスコンパイル
GOOS=linux GOARCH=amd64 go build -o skittles-linux .
GOOS=windows GOARCH=amd64 go build -o skittles.exe .
```

## ライセンス

Apache License 2.0

## クレジット

- [mint.x](https://opencode.ai) — X68000 用ファイラー。本プロジェクトはその設計思想に着想を得ています
- [bubbletea](https://github.com/charmbracelet/bubbletea) — TUI フレームワーク
- [lipgloss](https://github.com/charmbracelet/lipgloss) — スタイリング
- [glamour](https://github.com/charmbracelet/glamour) — Markdown レンダリング
- [chroma](https://github.com/alecthomas/chroma) — シンタックスハイライト
