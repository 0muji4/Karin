package ko

import (
	"math"
	"testing"
	"time"
)

// numberFromLongitude の境界を正確に検証する（天文計算に依存しない純粋な写像）。
func TestNumberFromLongitude(t *testing.T) {
	tests := []struct {
		lambda float64
		want   int
	}{
		{315.0, 1},  // 立春初候・東風解凍の起点
		{319.99, 1}, // バンド末尾
		{320.0, 2},  // 次のバンド
		{355.0, 9},  // 啓蟄末候
		{359.99, 9},
		{0.0, 10},  // 春分初候（315° 起点から 9 バンド進んで折り返す）
		{4.99, 10}, // バンド末尾
		{5.0, 11},
		{90.0, 28},  // 夏至初候・乃東枯の起点
		{180.0, 46}, // 秋分初候
		{270.0, 64}, // 冬至初候
		{310.0, 72}, // 大寒末候（立春の直前）
		{314.99, 72},
		{-5.0, 9},   // 負の入力も正規化されること（= 355°）
		{360.0, 10}, // 360° は 0° と同じ
	}
	for _, tt := range tests {
		if got := numberFromLongitude(tt.lambda); got != tt.want {
			t.Errorf("numberFromLongitude(%.2f) = %d, want %d", tt.lambda, got, tt.want)
		}
	}
}

// 全 360° を走査し、候が必ず 1..72 に収まり 5° ごとに切り替わることを確認する。
func TestNumberFromLongitude_rangeAndStep(t *testing.T) {
	for deg := 0; deg < 360; deg++ {
		ko := numberFromLongitude(float64(deg))
		if ko < Min || ko > Max {
			t.Fatalf("λ=%d° で候=%d が範囲外", deg, ko)
		}
	}
	// 立春(315°)から 5° ごとに 1→2→…→72→1 と進む。
	for band := 0; band < 72; band++ {
		lambda := math.Mod(315.0+float64(band)*5.0, 360.0)
		want := band + 1
		if got := numberFromLongitude(lambda + 1.0); got != want { // バンド内（+1°）
			t.Errorf("band %d (λ=%.0f) -> %d, want %d", band, lambda, got, want)
		}
	}
}

// 太陽黄経の天文計算が既知の基準と一致することを確認する。
func TestApparentSolarLongitude_anchors(t *testing.T) {
	tests := []struct {
		label  string
		t      time.Time
		target float64
		tol    float64
	}{
		{"J2000", time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), 280.37, 0.5},
		{"春分2026", time.Date(2026, 3, 20, 14, 46, 0, 0, time.UTC), 0.0, 0.5},
		{"夏至2026", time.Date(2026, 6, 21, 8, 24, 0, 0, time.UTC), 90.0, 0.5},
		{"秋分2026", time.Date(2026, 9, 23, 6, 5, 0, 0, time.UTC), 180.0, 0.5},
		{"冬至2026", time.Date(2026, 12, 21, 20, 50, 0, 0, time.UTC), 270.0, 0.5},
	}
	for _, tt := range tests {
		got := apparentSolarLongitudeDeg(tt.t)
		// 0° 近傍の折り返しを考慮した角度差。
		diff := math.Abs(math.Mod(got-tt.target+540, 360) - 180)
		if diff > tt.tol {
			t.Errorf("%s: λ=%.4f, want ~%.1f (±%.1f), diff=%.4f", tt.label, got, tt.target, tt.tol, diff)
		}
	}
}

// バンド中央の実日付で、候番号が暦どおりになることを確認する。
func TestNumber_knownDates(t *testing.T) {
	tests := []struct {
		label string
		t     time.Time
		want  int
	}{
		{"2026-02-08 立春の候", time.Date(2026, 2, 8, 3, 0, 0, 0, time.UTC), 1},    // 東風解凍
		{"2026-06-28 今日", time.Date(2026, 6, 28, 3, 0, 0, 0, time.UTC), 29},    // 菖蒲華
		{"2026-03-25 春分の候", time.Date(2026, 3, 25, 3, 0, 0, 0, time.UTC), 10},  // 雀始巣
		{"2026-12-25 冬至の候", time.Date(2026, 12, 25, 3, 0, 0, 0, time.UTC), 64}, // 乃東生
	}
	for _, tt := range tests {
		if got := Number(tt.t); got != tt.want {
			t.Errorf("%s: Number = %d, want %d", tt.label, got, tt.want)
		}
	}
}
