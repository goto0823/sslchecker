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

### 次にやること（ここから再開）

**B. 計算部分の切り出しとテスト**
- `notAfter` を受け取って残り日数を返す関数を作る
- `time.Now()` を関数の中で呼ばず外から渡す設計にすると、テストで任意の日付を渡せる
- `_test.go` を書いて `go test` を体験する。テーブル駆動テストの形を覚える
- ネットワーク不要の純粋計算なので、テスト入門に最適

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

## 既知の未対応事項

- タイムアウト未設定。応答しないホストで無限に待つ（`tls.Dialer` + `context`）
- `InsecureSkipVerify` 未使用のため、期限切れ証明書はエラーになり中身を読めない
- ポート 443 決め打ち、ホスト名もハードコード

## ユーザーが後でやりたいと言っていること

- TCP エコーサーバを書く（`net.Listen` + goroutine）。このツール完成後のご褒美
