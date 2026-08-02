package jobid

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNoTimestamp は、ジョブ ID から生成時刻を取り出せなかったことを表します。
// errors.Is で判定できます。
var ErrNoTimestamp = errors.New("job id has no embedded timestamp")

// timestampLayout はジョブ ID に埋め込まれる時刻の形式です。
const timestampLayout = "20060102150405"

// minTimestamp は、埋め込み時刻として妥当とみなす下限です。
//
// ジョブ ID には乱数由来の数字列が含まれるため、たまたま 14 桁の数字が並ぶことがあります。
// 月日や時分秒の範囲検査だけでは "0241-06-12 03:45:06" のような値を弾けないため、
// 現実にジョブが存在しえない時刻を除外します。
var minTimestamp = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

// CreatedAt は、ジョブ ID に埋め込まれた生成時刻を UTC で返します。
//
// 生成側（New）が UTC で採番するため、戻り値も UTC です。画面表示に使う場合は
// go-utils/jst で JST へ変換してください。
//
// New が生成する形式に加えて、New の導入前に各サービスが独自採番していた形式も
// 読み取れます。ジョブ ID はオブジェクトストレージ上の成果物のパスに使われており、
// 過去に採番された ID が残り続けるためです。対応する形式は次の 3 つです。
//
//	{prefix}-20060102-150405-{乱数}   New が生成する形式（prefix にハイフンを含んでよい）
//	c20060102-150405-{乱数}           プレフィックスが日付に直結する形式
//	20060102150405-{乱数}             日付と時刻が分割されない形式
//
// 時刻を取り出せない場合は ErrNoTimestamp をラップしたエラーを返します。
// 検証は行わないため、必要なら Validate と併用してください。
func CreatedAt(jobID string) (time.Time, error) {
	parts := strings.Split(jobID, "-")

	// 先頭側の要素を優先して走査し、最初に妥当と判定できた時刻を採用します。
	for i, part := range parts {
		// 日付と時刻が分割されていない形式。
		if t, ok := parseTimestamp(trimLeadingNonDigits(part)); ok {
			return t, nil
		}

		// 日付 (8 桁) と時刻 (6 桁) がハイフンで分かれている形式。
		// 先頭要素の非数字は "c20260803" のようなプレフィックス直結を吸収するために落とします。
		if i+1 < len(parts) {
			if date := trimLeadingNonDigits(part); len(date) == 8 && len(parts[i+1]) == 6 {
				if t, ok := parseTimestamp(date + parts[i+1]); ok {
					return t, nil
				}
			}
		}
	}

	return time.Time{}, fmt.Errorf("%w: %q", ErrNoTimestamp, jobID)
}

// SortKey は、ジョブ ID を作成日時で並べ替えるためのキーを返します。
// 時刻を取り出せない ID では空文字を返します。
//
// ID の文字列比較は、用途ごとに異なるプレフィックスを付けた ID が同じ一覧に混在すると
// プレフィックス順になってしまいます（時刻部分より前に差が出るため）。
// go-job-kit の paging.WithSortKey にそのまま渡せます。
func SortKey(jobID string) string {
	t, err := CreatedAt(jobID)
	if err != nil {
		return ""
	}
	return t.Format(timestampLayout)
}

// parseTimestamp は 14 桁の数字列を UTC の時刻として解釈します。
func parseTimestamp(value string) (time.Time, bool) {
	if len(value) != len(timestampLayout) || !isDigits(value) {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation(timestampLayout, value, time.UTC)
	if err != nil || t.Before(minTimestamp) {
		return time.Time{}, false
	}
	return t, true
}

// trimLeadingNonDigits は、先頭に付いた数字以外の文字を取り除きます。
func trimLeadingNonDigits(value string) string {
	return strings.TrimLeftFunc(value, func(r rune) bool {
		return r < '0' || r > '9'
	})
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
