package timeutil

import (
	"testing"
	"time"
)

// TestLoadLocationOrFallback_InvalidName は、存在しないタイムゾーン名を渡した場合に
// UTC+9 の FixedZone へフォールバックすることを確認します。
// JSTLocation は sync.Once でキャッシュされるためこの分岐を外部から直接検証できず、
// この内部テストで loadLocationOrFallback を単体テストしています。
func TestLoadLocationOrFallback_InvalidName(t *testing.T) {
	loc := loadLocationOrFallback("Not/A/Real/Zone")

	if loc.String() != "JST" {
		t.Errorf("フォールバック時のロケーション名が異なります。Got %s, want JST", loc.String())
	}

	_, offset := time.Now().In(loc).Zone()
	if offset != 9*60*60 {
		t.Errorf("フォールバック時のオフセットが異なります。Got %d, want %d", offset, 9*60*60)
	}
}

// TestLoadLocationOrFallback_ValidName は、正しいタイムゾーン名の場合に
// フォールバックせず、そのまま time.LoadLocation の結果を返すことを確認します。
func TestLoadLocationOrFallback_ValidName(t *testing.T) {
	loc := loadLocationOrFallback(jstLocationName)

	if loc.String() != jstLocationName {
		t.Errorf("ロケーション名が異なります。Got %s, want %s", loc.String(), jstLocationName)
	}
}
