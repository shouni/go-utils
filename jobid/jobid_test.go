package jobid

import (
	"errors"
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

	// 呼び出し側がエラーの種類で分岐できるよう、期待するエラーまで確認します。
	invalid := []struct {
		reason string
		id     string
		want   error
	}{
		{"空文字", "", ErrEmpty},
		{"先頭がハイフン", "-leading-hyphen", ErrInvalidFormat},
		{"先頭がアンダースコア", "_leading_underscore", ErrInvalidFormat},
		{"パス区切りを含む", "has/slash", ErrInvalidFormat},
		{"親ディレクトリ参照", "..", ErrInvalidFormat},
		{"パストラバーサル", "../etc/passwd", ErrInvalidFormat},
		{"空白を含む", "has space", ErrInvalidFormat},
		{"ドットを含む", "has.dot", ErrInvalidFormat},
		{"非 ASCII", "日本語", ErrInvalidFormat},
		{"長さ上限超過", strings.Repeat("a", MaxLength+1), ErrTooLong},
	}
	for _, tt := range invalid {
		if err := Validate(tt.id); !errors.Is(err, tt.want) {
			t.Errorf("Validate(%q) = %v, want %v (%s)", tt.id, err, tt.want, tt.reason)
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

// 空入力は path.Base が返す "." ではなく、入力そのものを表すエラーになること。
func TestSanitizeEmptyReportsMissingInput(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"", "   "} {
		_, err := Sanitize(input)
		if !errors.Is(err, ErrEmpty) {
			t.Errorf("Sanitize(%q) = %v, want %v", input, err, ErrEmpty)
		}
		if err != nil && strings.Contains(err.Error(), ".") {
			t.Errorf("Sanitize(%q) error mentions a path element that was not in the input: %v", input, err)
		}
	}
}

// プレフィックスが長すぎても ID 発行は失敗せず、切り詰められること。
func TestNewTruncatesLongPrefix(t *testing.T) {
	t.Parallel()

	id, err := New(strings.Repeat("a", MaxLength*2))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := Validate(id); err != nil {
		t.Fatalf("New() produced an invalid id %q: %v", id, err)
	}
	if len(id) != MaxLength {
		t.Fatalf("len(New()) = %d, want %d (the prefix should fill the id up to the limit)", len(id), MaxLength)
	}
}

// New が区切りのハイフンを付けるため、プレフィックス末尾の区切り文字が二重にならないこと。
func TestNewTrimsTrailingSeparatorInPrefix(t *testing.T) {
	t.Parallel()

	for _, prefix := range []string{"video-recipe-", "video-recipe__"} {
		id, err := New(prefix)
		if err != nil {
			t.Errorf("New(%q) error = %v", prefix, err)
			continue
		}
		if !strings.HasPrefix(id, "video-recipe-2") {
			t.Errorf("New(%q) = %q, want the separator collapsed", prefix, id)
		}
	}
}
