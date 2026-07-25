package jobid

import (
	"strings"
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	valid := []string{
		"20260725123456-abcd1234",               // ap-comp が生成する形式
		"video-recipe-20260725-150405-a1b2c3d4", // ap-mv が生成する形式
		"job_with_underscores",
		"a",
		strings.Repeat("a", MaxLength),
	}
	for _, id := range valid {
		if err := Validate(id); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", id, err)
		}
	}

	invalid := map[string]string{
		"":                               "空文字",
		"-leading-hyphen":                "先頭がハイフン",
		"_leading_underscore":            "先頭がアンダースコア",
		"has/slash":                      "パス区切りを含む",
		"..":                             "親ディレクトリ参照",
		"../etc/passwd":                  "パストラバーサル",
		"has space":                      "空白を含む",
		"has.dot":                        "ドットを含む",
		"日本語":                            "非 ASCII",
		strings.Repeat("a", MaxLength+1): "長さ上限超過",
	}
	for id, reason := range invalid {
		if err := Validate(id); err == nil {
			t.Errorf("Validate(%q) = nil, want an error (%s)", id, reason)
		}
	}
}

func TestSanitize(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"20260725123456-abcd1234":        "20260725123456-abcd1234",
		"  20260725123456-abcd1234  ":    "20260725123456-abcd1234",
		"prefix/20260725123456-abcd1234": "20260725123456-abcd1234",
		"../../20260725123456-abcd1234":  "20260725123456-abcd1234",
	}
	for input, want := range tests {
		got, err := Sanitize(input)
		if err != nil {
			t.Errorf("Sanitize(%q) error = %v", input, err)
			continue
		}
		if got != want {
			t.Errorf("Sanitize(%q) = %q, want %q", input, got, want)
		}
	}
}

// 末尾要素を取り出しても形式が不正なものは拒否すること。
func TestSanitizeRejectsInvalidBase(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"../../etc/pass wd", "", "dir/-leading", "dir/日本語", "/"} {
		if got, err := Sanitize(input); err == nil {
			t.Errorf("Sanitize(%q) = %q, want an error", input, got)
		}
	}
}

// Sanitize は末尾要素のみを取り出すため、末尾スラッシュは無視されます
// （path.Base の挙動）。ディレクトリ風の入力でも最後の要素が正当なら通ります。
func TestSanitizeUsesLastPathElement(t *testing.T) {
	t.Parallel()

	got, err := Sanitize("prefix/20260725123456-abcd1234/")
	if err != nil {
		t.Fatalf("Sanitize() error = %v", err)
	}
	if got != "20260725123456-abcd1234" {
		t.Fatalf("Sanitize() = %q", got)
	}
}

func TestNewProducesValidID(t *testing.T) {
	t.Parallel()

	id, err := New("video-recipe")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := Validate(id); err != nil {
		t.Fatalf("New() produced an invalid id %q: %v", id, err)
	}
	if !strings.HasPrefix(id, "video-recipe-") {
		t.Fatalf("New() = %q, want the prefix preserved", id)
	}
}

func TestNewIsUnique(t *testing.T) {
	t.Parallel()

	seen := make(map[string]struct{}, 100)
	for range 100 {
		id, err := New("job")
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("New() returned a duplicate id %q", id)
		}
		seen[id] = struct{}{}
	}
}

// 生成時刻を先頭に含めるため、辞書順のソートが新しい順の並びと一致すること。
// ストレージのキー順一覧をそのまま履歴の並びとして使えることを担保します。
func TestNewIsLexicographicallySortable(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 7, 25, 12, 34, 56, 0, time.UTC)
	older, err := newAt("job", base)
	if err != nil {
		t.Fatalf("newAt() error = %v", err)
	}
	newer, err := newAt("job", base.Add(time.Second))
	if err != nil {
		t.Fatalf("newAt() error = %v", err)
	}

	if older >= newer {
		t.Fatalf("expected %q < %q", older, newer)
	}
}

// プレフィックスの綴りで ID 発行が失敗しないこと。
func TestNewNormalizesPrefix(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"":               "job-",
		"   ":            "job-",
		"--":             "job-",
		"Regen Keyframe": "regenkeyframe-",
		"_private":       "private-",
		"日本語ジョブ":         "job-",
	}
	for prefix, wantPrefix := range tests {
		id, err := New(prefix)
		if err != nil {
			t.Errorf("New(%q) error = %v", prefix, err)
			continue
		}
		if !strings.HasPrefix(id, wantPrefix) {
			t.Errorf("New(%q) = %q, want prefix %q", prefix, id, wantPrefix)
		}
		if err := Validate(id); err != nil {
			t.Errorf("New(%q) produced an invalid id %q: %v", prefix, id, err)
		}
	}
}

func TestIsValid(t *testing.T) {
	t.Parallel()

	if !IsValid("20260725123456-abcd1234") {
		t.Error("IsValid() = false, want true")
	}
	if IsValid("../etc/passwd") {
		t.Error("IsValid() = true, want false")
	}
}
