// Package jobid は、非同期ジョブの識別子（ジョブ ID）の生成・検証・正規化を行います。
//
// ジョブ ID は HTTP のルートパラメータとオブジェクトストレージのパス要素の
// 両方に現れるため、検証はセキュリティ上の境界を兼ねます。複数のサービスが
// 同じジョブ ID をやり取りする構成では、「何を正当な ID とみなすか」が
// サービス間でずれていると、片方が発行した ID をもう片方が拒否したり、
// 一方だけがパストラバーサルを許してしまったりします。
// この境界を 1 箇所に集約するためのパッケージです。
package jobid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"
)

// MaxLength はジョブ ID に許容される最大文字数です。
const MaxLength = 128

// pattern は正当なジョブ ID の形式です。
//
// 先頭を英数字に限定しているのは、`-` や `_` で始まる値がコマンドライン引数や
// URL クエリで意図しない解釈をされるのを避けるためです。使用可能な文字を
// 英数字・ハイフン・アンダースコアに絞ることで、パス区切り (`/`)、親ディレクトリ
// 参照 (`..`)、URL エンコード文字を構造的に排除しています。
var pattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,` + fmt.Sprint(MaxLength-1) + `}$`)

// Validate は、ジョブ ID がルートおよびストレージパスで安全に扱える形式かを検証します。
func Validate(jobID string) error {
	if jobID == "" {
		return fmt.Errorf("job id is required")
	}
	if !pattern.MatchString(jobID) {
		return fmt.Errorf("invalid job id: %q", jobID)
	}
	return nil
}

// IsValid は Validate がエラーを返さないかどうかを返します。
func IsValid(jobID string) bool {
	return Validate(jobID) == nil
}

// Sanitize は、パス形式になりうる値を安全なジョブ ID へ正規化します。
//
// 外部から受け取った値をストレージパスへ組み込む前に通すことを想定しています。
// 末尾要素だけを取り出したうえで Validate を通すため、`../../etc/passwd` のような
// 入力は `passwd` に切り詰められ、形式が不正なら拒否されます。
// Validate だけでは「不正なら弾く」ことしかできませんが、Sanitize は
// 前段のルーティングや正規化で付いた余分なパス要素を落としてから判定します。
func Sanitize(jobID string) (string, error) {
	safe := path.Base(strings.TrimSpace(jobID))
	if err := Validate(safe); err != nil {
		return "", err
	}
	return safe, nil
}

// New は、指定されたプレフィックス付きのジョブ ID を生成します。
//
// 形式は `{prefix}-{yyyymmdd}-{hhmmss}-{ランダム 12 桁の 16 進数}` です。
// 先頭に UTC の生成時刻を含めるため、辞書順のソートがそのまま新しい順の
// 並び替えになります。オブジェクトストレージの一覧はキー順で返るため、
// この性質があると別途インデックスを持たずに済みます。
// prefix が空の場合は "job" を使います。
func New(prefix string) (string, error) {
	return newAt(prefix, time.Now().UTC())
}

// newAt は生成時刻を指定できる New です（テスト用）。
func newAt(prefix string, now time.Time) (string, error) {
	normalized := normalizePrefix(prefix)

	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("job id entropy: %w", err)
	}

	id := fmt.Sprintf("%s-%s-%s", normalized, now.Format("20060102-150405"), hex.EncodeToString(buf))
	if err := Validate(id); err != nil {
		return "", err
	}
	return id, nil
}

// normalizePrefix は、生成される ID が Validate を通るようにプレフィックスを整えます。
//
// プレフィックスは呼び出し側が用途を表すために付ける飾りであって、識別子の一意性は
// 時刻とランダム部分が担保します。そのため使えない文字が混ざっていてもエラーにはせず、
// 落として生成を続けます（プレフィックスの綴りのために ID 発行が失敗する方が困る）。
func normalizePrefix(prefix string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(prefix)) {
		if isAllowedInPrefix(r) {
			b.WriteRune(r)
		}
	}

	// 先頭が英数字でないプレフィックスは ID 全体を不正にしてしまうため取り除きます。
	normalized := strings.TrimLeft(b.String(), "_-")
	if normalized == "" {
		return "job"
	}
	return normalized
}

func isAllowedInPrefix(r rune) bool {
	return isAlphanumeric(r) || r == '-' || r == '_'
}

func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
