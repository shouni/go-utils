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
	"errors"
	"fmt"
	"path"
	"strings"
	"time"
)

// MaxLength はジョブ ID に許容される最大文字数です。
const MaxLength = 128

// Validate が返すエラーです。errors.Is で判定できます。
//
// 呼び出し側が「入力が無い」「長すぎる」「文字が不正」を区別できるようにするためのものです。
// たとえば HTTP のハンドラーではいずれも 400 になりますが、応答メッセージや
// メトリクスのラベルは分けたいことがあります。
var (
	// ErrEmpty は、ジョブ ID が空（または空白のみ）であることを表します。
	ErrEmpty = errors.New("job id is required")

	// ErrTooLong は、ジョブ ID が MaxLength を超えていることを表します。
	ErrTooLong = errors.New("job id is too long")

	// ErrInvalidFormat は、ジョブ ID に使えない文字が含まれることを表します。
	ErrInvalidFormat = errors.New("invalid job id")
)

// Validate は、ジョブ ID がルートおよびストレージパスで安全に扱える形式かを検証します。
//
// 正当な形式は「英数字で始まり、以降は英数字・ハイフン・アンダースコアのみ」で、
// 長さは MaxLength 文字までです。
//
// 先頭を英数字に限定しているのは、`-` や `_` で始まる値がコマンドライン引数や
// URL クエリで意図しない解釈をされるのを避けるためです。使用可能な文字を
// 英数字・ハイフン・アンダースコアに絞ることで、パス区切り (`/`)、親ディレクトリ
// 参照 (`..`)、URL エンコード文字を構造的に排除しています。
func Validate(jobID string) error {
	if jobID == "" {
		return ErrEmpty
	}
	if len(jobID) > MaxLength {
		return fmt.Errorf("%w: %d characters (max %d)", ErrTooLong, len(jobID), MaxLength)
	}

	// 許可する文字はすべて ASCII なので、バイト単位で走査できます。
	// マルチバイト文字は先頭バイトが許可集合から外れるため、この検査で弾かれます。
	if !isAlphanumeric(jobID[0]) {
		return fmt.Errorf("%w: %q", ErrInvalidFormat, jobID)
	}
	for i := 1; i < len(jobID); i++ {
		if !isAllowedInID(jobID[i]) {
			return fmt.Errorf("%w: %q", ErrInvalidFormat, jobID)
		}
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
// 前段のルーティングや正規化で付いた余分なパス要素を落としてから Validate を通すため、
// `../../etc/passwd` のような入力は `passwd` に切り詰められ、形式が不正なら拒否されます。
func Sanitize(jobID string) (string, error) {
	trimmed := strings.TrimSpace(jobID)

	// path.Base は空文字に "." を返すため、先に弾きます。そうしないと
	// 入力には無い "." がエラーメッセージに現れ、原因を取り違えさせます。
	if trimmed == "" {
		return "", ErrEmpty
	}

	safe := path.Base(trimmed)
	if err := Validate(safe); err != nil {
		return "", err
	}
	return safe, nil
}

// randomBytes は ID 末尾のランダム部分のバイト数です（16 進数で 2 倍の桁数になります）。
const randomBytes = 6

// generatedSuffixLength は New がプレフィックスの後ろに付ける部分の長さです。
const generatedSuffixLength = len("-20060102-150405-") + 2*randomBytes

// maxPrefixLength は、生成される ID が MaxLength に収まるプレフィックスの上限です。
const maxPrefixLength = MaxLength - generatedSuffixLength

// New は、指定されたプレフィックス付きのジョブ ID を生成します。
//
// 形式は `{prefix}-{yyyymmdd}-{hhmmss}-{ランダム 12 桁の 16 進数}` で、時刻は UTC です。
// prefix が空の場合は "job" を使います。
//
// 辞書順のソートがそのまま新しい順になるのは、一覧に並ぶ ID のプレフィックスが
// すべて同じ場合に限られます。オブジェクトストレージの一覧はキー順で返るため、
// その条件を満たすなら別途インデックスを持たずに済みます。プレフィックスが混在する
// 一覧では SortKey を並べ替えのキーに使ってください。
func New(prefix string) (string, error) {
	return newAt(prefix, time.Now().UTC())
}

// newAt は生成時刻を指定できる New です（テスト用）。
func newAt(prefix string, now time.Time) (string, error) {
	normalized := normalizePrefix(prefix)

	buf := make([]byte, randomBytes)
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
// 長すぎる場合も同じ理由で、エラーにせず maxPrefixLength まで切り詰めます。
// 切り詰めた結果 ID が MaxLength より短くなることはありますが、一意性は
// 時刻とランダム部分が担保するため問題になりません。
func normalizePrefix(prefix string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(prefix)) {
		if r < 0x80 && isAllowedInID(byte(r)) {
			b.WriteByte(byte(r))
		}
	}

	// 先頭が英数字でないプレフィックスは ID 全体を不正にしてしまうため取り除きます。
	// 末尾の区切り文字も落とします。New が区切りのハイフンを付けるため、残すと
	// "prefix--20060102-..." のように二重になります。
	normalized := strings.Trim(b.String(), "_-")
	if normalized == "" {
		return "job"
	}

	// 残るのは ASCII だけなので、バイト単位で切っても文字は壊れません。
	// 切った位置が区切り文字に当たることがあるため、ここでも末尾を落とします。
	if len(normalized) > maxPrefixLength {
		normalized = strings.TrimRight(normalized[:maxPrefixLength], "_-")
	}
	return normalized
}

func isAllowedInID(c byte) bool {
	return isAlphanumeric(c) || c == '-' || c == '_'
}

func isAlphanumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
