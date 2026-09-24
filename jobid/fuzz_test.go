package jobid

import (
	"strings"
	"testing"
	"time"
)

// FuzzSanitize は、Sanitize が返す値が必ず Validate を通ることを確かめます。
//
// 検証はセキュリティ境界です。ジョブ ID は URL のパスにもストレージのパスにも現れ、
// Sanitize は外部から来た値をそこへ組み込む前の最後の関門です。ここが「不正なのに
// 通った」を一度でも許すと、パス区切りや親ディレクトリ参照がそのまま下流へ渡ります。
// 表のテストは思いついた形しか試せないので、探索で埋めます。
func FuzzSanitize(f *testing.F) {
	seeds := []string{
		"20260725123456-abcd1234",
		"music-20260725-123456-abcdef123456",
		"../../etc/passwd",
		"a/b/c",
		"  spaced  ",
		"-leading-hyphen",
		"_leading-underscore",
		".",
		"..",
		"/",
		"",
		"あいうえお",
		strings.Repeat("a", MaxLength),
		strings.Repeat("a", MaxLength+1),
		"a\x00b",
		"a%2Fb",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		safe, err := Sanitize(raw)
		if err != nil {
			if safe != "" {
				t.Fatalf("Sanitize(%q) failed but returned %q", raw, safe)
			}
			return
		}

		// 成功したなら、その値は必ず Validate を通る。
		if err := Validate(safe); err != nil {
			t.Fatalf("Sanitize(%q) = %q, which Validate rejects: %v", raw, safe, err)
		}

		// 危険な形が残っていないこと。Validate の文字集合から導かれる帰結だが、
		// 境界そのものなので直接も確かめる。
		if strings.ContainsAny(safe, `/\.%`+"\x00") || strings.TrimSpace(safe) != safe {
			t.Fatalf("Sanitize(%q) = %q, which is not path-safe", raw, safe)
		}
		if len(safe) > MaxLength {
			t.Fatalf("Sanitize(%q) = %q (%d bytes), over MaxLength", raw, safe, len(safe))
		}

		// 冪等であること。通った値をもう一度通しても変わらない。
		again, err := Sanitize(safe)
		if err != nil || again != safe {
			t.Fatalf("Sanitize is not idempotent: %q -> %q -> (%q, %v)", raw, safe, again, err)
		}
	})
}

// FuzzCreatedAt は、CreatedAt が成功したときの時刻が妥当で、SortKey と一致することを
// 確かめます。
//
// 解析対象は「過去に採番されて今もストレージに残っている 3 つの形式」で、ハイフンで
// 割った要素を前から舐めるという手書きの走査です。一覧の並びはこの戻り値で決まるため、
// 取り違えると履歴の順序が静かに崩れます。
func FuzzCreatedAt(f *testing.F) {
	seeds := []string{
		"20260725123456-abcd1234",
		"c20260803-101112-abcd1234",
		"music-20260725-123456-abcdef123456",
		"video-gen-20260725-123456-abcdef123456",
		"no-timestamp-here",
		"19990101000000-old",
		"20261332999999-invalid",
		"-------",
		"",
		"12345678-901234",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		got, err := CreatedAt(raw)
		key := SortKey(raw)

		if err != nil {
			if key != "" {
				t.Fatalf("CreatedAt(%q) failed but SortKey returned %q", raw, key)
			}
			return
		}

		// 採番が始まる前の時刻は、解析の取り違えとしてしか出てこない。
		if got.Before(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("CreatedAt(%q) = %v, before the minimum", raw, got)
		}
		if got.Location() != time.UTC {
			t.Fatalf("CreatedAt(%q) = %v, want UTC", raw, got)
		}
		// 並べ替えキーは同じ時刻から導かれる。
		if want := got.Format(timestampLayout); key != want {
			t.Fatalf("SortKey(%q) = %q, want %q", raw, key, want)
		}
	})
}
