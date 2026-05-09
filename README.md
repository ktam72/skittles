# Skittles

X68000 用ファイラー [mint.x](https://opencode.ai) の設計思想を継承した、
モダンな 2画面 TUI ファイルマネージャ。

![Skittles Screenshot](screens/screen1.png)
![Skittles Screenshot](screens/screen2.png)
![Skittles Screenshot](screens/screen3.png)

## 特徴

- **2画面 + コンソール の3ペイン**: Tab でフォーカス切替、アクティブペインが source、反対が destination
- **ファイル操作**: コピー/移動/削除（確認ダイアログ）/マーク/ソート/リネーム
- **パーミッション・所有者・グループ表示**: 各行に `drwxr-xr-x owner group` 形式
- **ビルトインビューア**: テキスト・Markdown・ソースコードのシンタックスハイライト対応（スクロール可能）。**バイナリは自動HEX表示**
- **文字コード自動判別**: UTF-8 / Shift-JIS / EUC-JP を自動検出して表示
- **アーカイブ内部ブラウズ**: ZIP/TAR/7z/GZ/BZ2 を **pure Goで内部展開**、RARはpure Go + unarフォールバックでディレクトリのように閲覧・操作
- **拡張子→アクション**: `.go` → ビューア、`.zip` → ブラウズ、`.mdx` → MP4M.app 等
- **コンソールペイン**: **cd対応・`$`プロンプト表示**、コマンドの**リアルタイム出力**、履歴
- **クロスプラットフォーム**: macOS / Linux / Windows（pure Go、CGo不要）
- **日本語入力対応**: macOS 起動時・コンソールフォーカス時に自動で英数入力に切替
- **トップバー**: アプリ名・バージョン・クレジット・リアルタイム時計（曜日付き）表示
- **外部プレビュー**: `p` キーで外部アプリ（プレビュー.app等）で開く

## インストール

### Homebrew（準備中）

```bash
brew install ktam72/tap/skittles
```

### Go

```bash
go install github.com/ktam72/skittles/src@latest
```

※ モジュールルートが `src/` のため `.../skittles/src` を指定します。

### 手動ビルド

```bash
git clone git@github.com:ktam72/skittles.git
cd skittles/src

# ビルド（CGo不要）
go build -o ../skittles .

# 実行
cd .. && ./skittles
```

#### アーカイブ形式対応

| 形式 | エンジン |
|------|---------|
| ZIP | Go標準 `archive/zip` |
| TAR/TGZ | Go標準 `archive/tar` + `compress/gzip` |
| BZ2/TBZ2 | Go標準 `compress/bzip2` |
| 7z | `github.com/bodgit/sevenzip`（pure Go） |
| RAR | `github.com/nwaples/rardecode/v2`（pure Go、読み取り専用） |

全て pure Go。CGo不要。外部コマンド依存ゼロ。

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
Space → カーソル行をマーク（複数選択）
    a → 全ファイルをマーク
    c → マークしたファイルを反対ペインへコピー
    m → マークしたファイルを反対ペインへ移動
    d → 削除（確認ダイアログ表示）
    r → ファイル名リネーム
```

マークがない場合、コピー/移動/削除はカーソル上の1ファイルのみが対象になります。

#### アーカイブ（ZIP/TAR/7z等）の操作

```
  Enter → アーカイブ内部をブラウズ（ディレクトリのように表示）
.. / ← / Backspace → アーカイブルートでは元のディレクトリに戻る
                      （サブフォルダ内では親フォルダへ）
    x → アーカイブをカレントディレクトリに展開
（アーカイブブラウズ中は枠・文字色がピンク色に変化します）
```

アーカイブ内部では通常のファイル操作（コピー・移動・削除）がそのまま使えます。
アーカイブから戻ると、元のアーカイブファイルにカーソルが復帰します。

#### コンソールペイン

```
  ! → コンソールにフォーカス移動
       cd コマンドが使えます（コンソール専用のカレントディレクトリを持つ）
       コマンドを入力 → Enter で実行（結果は goroutine でリアルタイム表示）
  ↑↓ → コマンド履歴（または出力スクロール）
 ESC → ファイルペインに戻る
```

#### ビルトインビューア

```
   Enter → テキスト/ソースコード/HEX をインライン表示
   ↑↓   → 1行スクロール
 ←/→/PgUp/PgDown → ページスクロール
   ESC  → 閉じる
```

Markdown（`.md`）は色付きでレンダリング、ソースコードは chroma でシンタックスハイライト、
バイナリは自動で HEX 表示されます。文字コードは UTF-8 / Shift-JIS / EUC-JP を自動判別。

#### その他

```
   p → 外部アプリでプレビュー（画像はプレビュー.app等）
   E → エディタ（$EDITOR）で開く
   R → ディレクトリ再読込
  sr → ソート順切替（名前→日時→拡張子→サイズ）
 ESC → ESC×2 で終了
```

### キーバインド一覧

| キー | 機能 |
|------|------|
| `↑↓ / kj` | カーソル移動 |
| `→ / l` | ブラウズモードでは無効 |
| `Enter` | ディレクトリ進入 / ファイルを開く / 実行ファイルはコンソールにパス入力 |
| `← / h / Backspace` | 親ディレクトリへ / アーカイブを抜ける |
| `Tab` | フォーカス切替（Left→Right→Console→Left） |
| `PgUp / PgDown / b` | ページ送り / 戻し |
| `Space` | マーク |
| `a` | 全マーク |
| `p` | 外部アプリでプレビュー |
| `c` | 反対ペインへコピー |
| `m` | 反対ペインへ移動 |
| `d` | 削除（確認ダイアログ表示） |
| `r` | ファイル名リネーム |
| `R` | カレントディレクトリ再読込 |
| `sr` | ソート切替 |
| `!` | コンソールへジャンプ |
| `E` | エディタで開く |
| `x` | アーカイブをカレントディレクトリに展開 |
| `ESC` | 終了（2回押し） |

### ビューア操作

| キー | 機能 |
|------|------|
| `↑↓` | 1行スクロール |
| `←→ / PgUp / PgDown` | 1ページスクロール |
| `ESC` | 閉じる |

## 設定

`config.yaml` で拡張子ごとのアクションをカスタマイズできます。

```yaml
actions:
  - match: ".md"
    viewer: true
  - match: ".zip"
    browse: true
  - match: ".mdx"
    command: "open -a MP4M.app $P"
```

変数: `$P` = フルパス、`$F` = ファイル名、`$D` = ディレクトリ、`$EDITOR` = エディタ

## 開発

```bash
cd src
go build -o ../skittles .              # ビルド（CGo不要）
go vet ./...                            # 静的解析
golangci-lint run ./...                 # Lint（0 issues）

# クロスコンパイル
GOOS=linux GOARCH=amd64 go build -o ../skittles-linux .
GOOS=windows GOARCH=amd64 go build -o ../skittles.exe .
```

## ライセンス

Apache License 2.0

## クレジット

- [mint.x](https://opencode.ai) — X68000 用ファイラー。本プロジェクトはその設計思想に着想を得ています
- [bubbletea](https://github.com/charmbracelet/bubbletea) — TUI フレームワーク
- [lipgloss](https://github.com/charmbracelet/lipgloss) — スタイリング
- [glamour](https://github.com/charmbracelet/glamour) — Markdown レンダリング
- [chroma](https://github.com/alecthomas/chroma) — シンタックスハイライト
- [sevenzip](https://github.com/bodgit/sevenzip) — 7z アーカイブリーダー
