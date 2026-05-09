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
- **ビルトインビューア**: テキスト・Markdown・ソースコードのシンタックスハイライト対応（スクロール可能）。**バイナリは自動HEX表示**
- **アーカイブ内部ブラウズ**: ZIP/TAR/7z/LZH/RAR/GZ を **ディレクトリのように閲覧・操作**
- **拡張子→アクション**: `.go` → ビューア、`.zip` → ブラウズ、`.mdx` → MP4M.app 等
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
cd skittles/src

# 依存パッケージ（CGoビルドに必要）
brew install libarchive p7zip

# ビルド
go build -o ../skittles .

# 実行
cd .. && ./skittles
```

#### アーカイブ形式対応と依存

| 形式 | CGo有効（推奨） | CGo無効（`CGO_ENABLED=0`） |
|------|----------------|---------------------------|
| ZIP/TAR/GZ | Homebrew libarchive | Go標準ライブラリ（依存不要） |
| 7z/LZH/RAR | **p7zip**（フォールバック） | 外部コマンド（`7z`/`lha`/`unrar`） |
| BZ2 | Homebrew libarchive | 外部コマンド（`bunzip2`） |

CGo無効の場合は `brew install p7zip lha unrar` 等が必要です。

## 使い方

```bash
# カレントディレクトリで起動
./skittles

# 左右別のディレクトリで起動
./skittles /usr/local /opt/homebrew

# 同じディレクトリで起動
./skittles ~/Documents
```

### 基本操作

**3ペイン構成**: 左ペイン・右ペイン・コンソールペイン。Tab でフォーカスを切り替えます。
アクティブなペインが **source**、反対のファイルペインが **destination** として機能します。

#### ファイルの選択と操作

```
Space (or Backspace) → カーソル行をマーク（複数選択）
                a → 全ファイルをマーク
                c → マークしたファイルを反対ペインへコピー
                m → マークしたファイルを反対ペインへ移動
                d → 削除（確認ダイアログ表示）
```

マークがない場合、コピー/移動/削除はカーソル上の1ファイルのみが対象になります。

#### アーカイブ（ZIP/TAR/7z等）の操作

```
  Enter → アーカイブ内部をブラウズ（ディレクトリのように表示）
   ←   → ブラウズ終了、元のディレクトリに戻る
    x   → アーカイブをカレントディレクトリに展開
（アーカイブブラウズ中は枠・文字色がピンク色に変化します）
```

アーカイブ内部では通常のファイル操作（コピー・移動・削除）がそのまま使えます。

#### コンソールペイン

```
  !  → コンソールにフォーカス移動
        コマンドを入力 → Enter で実行（結果はリアルタイム表示）
  ↑↓ → コマンド履歴
 ESC → ファイルペインに戻る
```

#### ビルトインビューア

```
   Enter → テキスト/ソースコード/HEX をインライン表示
   ↑↓   → 1行スクロール
 ←/→/PgUp/PgDown → ページスクロール
   ESC  → 閉じる
```

Markdown（`.md`）は色付きでレンダリング、ソースコードはシンタックスハイライト、
バイナリファイルは自動で HEX 表示されます。

#### その他

```
   p  → 外部アプリでプレビュー（画像はプレビュー.app等）
   E  → エディタ（$EDITOR）で開く
   sr → ソート順切替（名前→日時→拡張子→サイズ）
   r  → ディレクトリ再読込
 ESC → ESC×2 で終了
```

### キーバインド一覧

| キー | 機能 |
|------|------|
| `↑↓ / kj` | カーソル移動 |
| `→ / l` | ブラウズモードでは無効 |
| `Enter` | ディレクトリ進入 / ファイルを開く / 実行ファイルはコンソールにパス入力 |
| `← / h` | 親ディレクトリへ |
| `Tab` | フォーカス切替（Left→Right→Console→Left） |
| `Space / b` | ページ送り / 戻し |
| `Space / Backspace` | マーク |
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

> **MacのBackspace**: `delete` キー（Returnキーの上）が Backspace として動作します。
> Space キーでもマークできるので、そちらもご利用ください。

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
# 依存（CGo有効時）
brew install libarchive p7zip

# ビルド（src/ 内で実行）
cd src
CGO_ENABLED=1 go build -o ../skittles .
cd ..
go vet ./src/...                     # 静的解析
golangci-lint run ./src/...          # Lint（0 issues）

# 非CGoビルド（外部コマンド依存）
cd src && CGO_ENABLED=0 go build -o ../skittles .

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
