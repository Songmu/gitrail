---
created: 2026-03-12 21:49
modified: 2026-03-15 14:04
---

## 概要

Git履歴を辿って、ある期間におけるファイルの変更を表示するCLI「**gitrail**」。gitコマンドのラッパとしてGoで作る。

## 仕様

### CLIインターフェース

コマンド名: **gitrail**

```bash
gitrail --since="2026-01-01" --until="2026-03-01" [-C /path/to/repo] [--branch=main] [--json] [-- pathspec...]
```

| フラグ              | デフォルト      | 説明                                           |
| ---------------- | ---------- | -------------------------------------------- |
| `-C`             | カレントディレクトリ | 対象Gitリポジトリのパス                                |
| `--since`        | **必須**     | 開始時刻                                         |
| `--until`        | **必須**     | 終了時刻                                         |
| `--branch`       | HEAD       | 対象ブランチ（任意のrevisionも受け付ける）                    |
| `--json`         | false      | NDJSON出力                                     |
| `-- pathspec...` | なし         | Gitのパス指定構文（複数指定可。例: `'*.go'`, `':!vendor/'`） |

- 時刻フォーマットはGitに委譲する。ISO 8601（`2026-01-01`）、相対指定（`"1 month ago"`）等がそのまま使える
- タイムゾーン省略時はローカルタイムゾーンとして扱われる（Gitの挙動）
- flagパースには標準`flag`パッケージを使用。`--` 以降の引数は `fs.Args()` でpathspecとして取得

### 処理の流れ

1. 起点のコミットハッシュの取得
	- ブランチ 上の 開始時刻 直前のコミット / 無ければ直後のコミット
	- `git log --first-parent --before="<開始時刻>" -1 --format="%H" <ブランチ>`
	- 無い場合: `git log --first-parent --after="<開始時刻>" --reverse --format="%H" <ブランチ>` の出力から先頭行を取得
		- 注: `-n 1` と `--reverse` を併用すると、`-n 1` が `--reverse` より先に適用され最新のコミットが返されるため、`-n 1` は使わず出力の先頭行を取る
2. 終点のコミットハッシュの取得
	- ブランチ 上の 終了時刻 直前のコミット
	- `git log --first-parent --before="<終了時刻>" -1 --format="%H" <ブランチ>`
3. 起点と終点が同一コミットの場合は差分なしとして早期リターン
4. 起点が終点の祖先であることを検証
	- `git merge-base --is-ancestor <起点> <終点>`
5. 起点と終点に対してディレクトリの `git diff` を取って変更があったファイル一覧を抽出する
	- 追加・変更・削除
	- `git diff --name-status <起点hash> <終点hash> -- <pathspec>...`
6. 5で取得されたファイルに対して、リネームを極力検知する
	- 起点〜終点間のコミットを走査し、各コミットのリネーム情報を収集する
	- `git log --first-parent -M --diff-filter=R --name-status --format="%H" <起点>..<終点> -- <pathspec>...`
	- 各コミット単位では差分が小さいため `-M` の精度が高く、リネーム後に大幅変更があっても追跡可能
	- リネームチェーン（A→B→Cのような連続リネーム）の構築も可能
		- git logの出力（新しい順）を逆順（古い順）に処理し、originMapを構築する
	- リネーム後に元のファイルと同じ名前のファイルが作られるケースも、コミット単位で追えば検知できる
	- 起点に存在するファイルのみをリネーム追跡する
		- 起点に存在しないファイルが途中で新規追加され、その後リネームされたケースについては、リネームは無視して終点における新規追加ファイルとして扱う（起点に元ファイルが無いためリネーム追跡の意味がない）
		- 具体的には、diff結果のDeletedセット（起点に存在）とAddedセット（終点に存在）の両方にマッチするリネームのみ適用する
7. 結果をパス名のアルファベット順でソートして返却

### ライブラリインターフェース

CLIとは別に、Goライブラリとしても利用可能にする。

```go
// Gitrail は設定を保持する構造体
type Gitrail struct {
    Dir       string    // 対象Gitリポジトリのパス。空の場合はカレントディレクトリ
    ErrStream io.Writer // gitエラー出力先。nilの場合はos.Stderr
}

// New はコンストラクタ
func New(dir string) *Gitrail

// Trail はブランチ上のsince〜until間のファイル変更を返す
// pathspecsはオプション（可変長引数）
func (g *Gitrail) Trail(ctx context.Context, branch string, since, until time.Time, pathspecs ...string) (*Result, error)
```

- `Gitrail` 構造体がリポジトリパスとエラー出力先を保持し、`Trail` メソッドでブランチ・期間・pathspecsを受け取る
- CLIではユーザー入力の文字列をそのままGitに渡す（相対指定等も含め柔軟に対応）
- ライブラリでは `time.Time` を受け取り、内部で `time.RFC3339` にフォーマットしてGitコマンドに渡す

### 出力データ構造

```go
type Result struct {
    StartCommit string       // 起点コミットハッシュ
    EndCommit   string       // 終点コミットハッシュ
    Changes     []FileChange // パス名のアルファベット順でソート済み
}

type FileChange struct {
    Status  ChangeStatus // Added | Modified | Deleted
    Path    string       // 終点時点のファイルパス。Deletedの場合は起点時点のパス
    OldPath string       // リネーム元パス。リネームがあった場合のみ設定
}

type ChangeStatus string

const (
    Added    ChangeStatus = "Added"
    Modified ChangeStatus = "Modified"
    Deleted  ChangeStatus = "Deleted"
)
```

- ステータスは3種のみ。Gitの Renamed/Copied/TypeChanged 等は扱わない
	- リネームはOldPathの有無で表現する
	- Copied は Added として扱う
	- TypeChanged は Modified として扱う

### 出力形式

#### テキスト形式（デフォルト）

1行目にcommit range、空行で区切ってタブ区切りのファイル一覧。リネームがある場合は3カラム目にリネーム元パス。
差分がない場合（起点・終点が同一コミットの場合も含む）はcommit rangeのみ出力。
ファイル一覧はパス名のアルファベット順でソートする。

```
abc123..def456

A	src/new.go
M	src/foo.go
M	src/bar.go	src/old_bar.go
D	src/removed.go
```

差分なしの場合:
```
abc123..def456
```

#### NDJSON形式（`--json`）

1行1ファイル。各行にcommit hashを含め、行単体で独立して解釈可能にする。
差分がない場合は出力なし（0行）。

```jsonl
{"end_commit":"def456","status":"Added","path":"src/new.go"}
{"start_commit":"abc123","end_commit":"def456","status":"Modified","path":"src/foo.go"}
{"start_commit":"abc123","end_commit":"def456","status":"Modified","path":"src/bar.go","old_path":"src/old_bar.go"}
{"start_commit":"abc123","status":"Deleted","path":"src/removed.go"}
```

### エラーハンドリング

判定はコミット取得結果ベースで行う:

1. 起点のコミットが見つからない場合（フォールバック含め。空リポジトリ等）: **終了コード1**（素のerror）
2. 終点のコミットが見つからない場合（履歴の範囲外）: **終了コード2**（`ExitCode() int` インターフェース付きerror）。stderrにメッセージ出力
	- 注: `until < since < first_commit` のケースも終了コード2になるが、コミット取得結果からは区別できないため許容する
3. 両方見つかったが逆転している場合（起点が終点の祖先でない）: **終了コード1**（素のerror）
	- `git merge-base --is-ancestor <起点> <終点>` で検証
- 対象ディレクトリがGitリポジトリでない場合: **終了コード1**
- 内部で実行するgitコマンドが非0で終了した場合（不正なブランチ名など）: **終了コード1**。Gitのエラーメッセージをstderrにそのまま出力

終了コード1はmain.goのデフォルト終了コードであるため、素の `error` を返せばよい。終了コード2のみ `ExitCode() int` インターフェースを実装した `exitError` 型を使う。

#### 終了コード

| コード | 意味                               |
| --- | -------------------------------- |
| 0   | 成功（差分あり・なし問わず）                   |
| 1   | エラー（リポジトリでない、コミット逆転、起点が見つからないなど） |
| 2   | 終点のコミットが見つからない（履歴の範囲外）           |

### 補足

- コミットハッシュの取得を一律直前のコミットとしているのは、コミット時刻から指定時刻の間のファイル状態は不定であるが指定時刻における状態を担保するのが直前のコミットしか無いため
- Gitコマンド実行時は `LANG=C`, `LC_ALL=C` を環境変数に設定し、出力ロケールを固定する
- Gitコマンドのstderrはユーザーに見せるため、errStreamに接続する
- 手順6のリネーム検知におけるパフォーマンス懸念
	- コミット数に比例して処理時間が増加する。問題があれば下記の代替案を検討
	- **代替案: `git diff -M` による2点間比較**
		- `git diff -M --name-status <起点hash> <終点hash> -- <pathspec>...` で step 5 と統合可能
		- 起点と終点のファイル内容の類似度のみで判断するため、リネーム後に大幅な変更があると検知できない弱点がある
		- ただし1コマンドで完結しコミット数に依存せず高速

