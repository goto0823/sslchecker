# sslChecker

Go の学習を兼ねた、複数の SSL 証明書の有効期限をチェックする CLI ツール。

## このプロジェクトでの進め方（最重要）

**コードは書かないでください。** 学習が目的なので、実装は必ずユーザー自身が書きます。

- 実装（関数の中身）を提示しない。書き写せる形のコードを出さない
- 型・シグネチャの「形」、API 名、標準ライブラリの関数名を伝えるのは可
- ユーザーが書いたコードを `go vet` / `go run` で実際に動かして確認し、指摘する
- 指摘は必ず理由とセットで。「なぜそうするか」を Go の文化・慣習とともに説明する
- 答えを先回りして全部言わない。まず自分で書かせ、詰まったらヒントを段階的に出す
- コンパイルエラーは「読み方」から説明する。エラーメッセージを教材として使う

## 学習方針

- pkg.go.dev と `go doc` を自分で引く習慣をつけてもらう
- 口で説明するより、実際にコマンドを動かして見せるほうが理解が早い
- 標準ライブラリのソースを読むことを推奨する（`go doc -src`）
- 寄り道は楽しいが沼るので、まずツールを完成させる方向に戻す

## 環境

- Nix flake + direnv で Go 1.27（`flake.nix` にピン留め済み）
- 確認コマンド: `go vet ./...` と `go run ./cmd/sslchecker`
- モジュール名が `sslCheker`（Checker のタイポ）。直すなら go.mod の1行
- `.gitignore` で `/.go/`（GOPATH。モジュールキャッシュが入る）と `/sslchecker`（ビルド成果物）を除外済み

## 進捗

### 完了

**Step 1: 1ホストの証明書取得と残り日数の計算**
- `fetchCert(host string) (*x509.Certificate, error)` に切り出し済み
- `net.JoinHostPort` でホストとポートを結合（IPv6 対応のため）
- `defer conn.Close()` は err チェックの直後
- `len(peerCerts) == 0` のガードを添字アクセスの前に

**エラー処理**
- `fmt.Errorf` + `%w` でラップ。書式は `fetch cert <host>: <原因>`
- エラーは `os.Stderr` へ、失敗時は `os.Exit(1)`

**Step 2: 残り日数の計算を切り出し**
- `daysUntil(notAfter, now time.Time) int` に切り出し済み
- `time.Now()` は `main` 側で呼び、引数で渡す設計（テストで任意の日付を渡せる）

**テーブル駆動テストの土台（`cmd/sslchecker/main_test.go`）**
- `TestDaysUntil` が匿名構造体スライス + `t.Run` のサブテスト形式で動作中。緑
- `now` は `time.Date(2026, 9, 3, 17, 15, 0, 0, time.UTC)` で固定（UTC 固定は実行環境非依存にするため）
- フィールドは `name` / `d time.Duration` / `want int`。`notAfter` はループ内で `now.Add(tt.d)` で組み立てる
- 現在のケースは「30日」1件のみ

### 次にやること（ここから再開）

**B-2. テストケースを増やす（ここから）**
- 追加してもらう予定のケース: 端数あり（30日と12時間）/ ちょうど 0 / マイナス12時間 / マイナス24時間
- `want` は「実装がこう返すはず」ではなく「仕様としてこうあるべき」で書いてもらう方針を伝え済み
- **仕込み済みの論点（まだ答えを言っていない）**: `daysUntil` の `int(d.Hours() / 24)` は
  ゼロ方向への切り捨てであって floor ではない。負の Duration だと「12時間前に失効」が 0 になり、
  「あと数時間ある」と区別できない。テストが落ちてから
  「テストが間違っているのか実装が間違っているのか」を自分で考えてもらう段階
- ケースが出揃ったら `daysUntil` の仕様（切り捨ての向き、失効済みの表現）を決めて実装を直す

**C. 複数ホスト対応**
- まず `for` で逐次 → その後 goroutine + `sync.WaitGroup` で並行化
- 並行数の制限が実用上必須（バッファ付き channel をセマフォに、または `errgroup.SetLimit`）
- `go run -race` / `go test -race` で確認

**その後**
- CLI 化（`flag`、ファイル/標準入力からホスト一覧、`text/tabwriter`、JSON 出力）
- テスト（`httptest.NewTLSServer`、`x509.CreateCertificate` で期限切れ証明書を自作）
- 発展: 中間証明書の期限、Nagios 互換の終了コード（0=OK/1=WARN/2=CRIT）、`run() error` パターン

## 説明済みの事項（再説明は不要、参照はOK）

- TLS ハンドシェイクと証明書チェーン。`PeerCertificates[0]` がリーフである理由
- `go doc` と pkg.go.dev の使い分け。メソッドは型のセクションの下に並ぶ
- `Close()` と fd リーク。`defer` の位置（err チェックの後）
- `io.Closer` / `net.Conn` / `tls.Conn` の層構造
- `error` はインターフェース。`%w` によるラップ、`errors.Is` / `errors.As`
- `fmt` の Print / Sprint / Fprint / Errorf 一族と接尾辞 f・ln の意味
- stdout と stderr の使い分け、リダイレクト、終了コード（0=成功）
- 命名規則（名前に型を入れない、`Get` を使わない、I/O を伴うなら動詞）
- `time.Time` と `time.Duration`、`Hour()` と `Hours()` の違い、`Days()` が無い理由
- タイムゾーンは表示だけの問題。残り日数の計算には影響しない
- `x509` のフィールド名は RFC 5280 の用語（`NotBefore` / `NotAfter`）
- `go test` はテスト関数を命名規則で発見する（`TestXxx`、`Test` の次の文字は小文字不可）。
  名前が違うと沈黙して `no tests to run` になる。`go test` は裏で `go vet` の一部を実行する
- 未使用でコンパイルエラーになるのはローカル変数と import だけ。パッケージレベルの関数・変数は対象外
- 関数内のローカル変数のスコープは宣言地点から。パッケージレベルの宣言は順序不問（この非対称性）
- 先頭が `0` の数値リテラルは8進数。`09` はエラー、`010` は黙って 8 になる。日時をゼロパディングしない
- `time.Hour`（`Duration` 定数、括弧なし）と `now.Hour()`（メソッド、括弧あり）の違い。
  `Duration` の実体は int64 のナノ秒
- `go doc` の読み方: 宣言が `const`/`var`/`func`/`type` のどれで始まるか。
  関数は戻り値の型の下にグループ化される（`time.Date` が `type Time` の下にいる理由）
- Go の定数は ALL_CAPS にしない。名前からは定数だと判断できない
- 命名は「スコープの広さに名前の長さを比例させる」。標準ライブラリの `d Duration` / `u Time` が実例
- 複合リテラルは最後の要素にも末尾カンマが必要（セミコロン自動挿入の仕様）
- テーブルに入れるのは「ケースごとに結果を変える軸」だけ。`now` のような固定値は関数の先頭へ
- `t.Errorf` は `got, want` の順。`go test -run 'TestXxx/サブテスト名'` で1ケースだけ実行できる

## 既知の未対応事項

- タイムアウト未設定。応答しないホストで無限に待つ（`tls.Dialer` + `context`）
- `InsecureSkipVerify` 未使用のため、期限切れ証明書はエラーになり中身を読めない
- ポート 443 決め打ち、ホスト名もハードコード

## ユーザーが後でやりたいと言っていること

- TCP エコーサーバを書く（`net.Listen` + goroutine）。このツール完成後のご褒美
