package ko

import (
	"math"
	"testing"
	"time"
)

// KoStart は t 以前で、t の属する候の開始境界（λ が 5° の倍数）を返す。
func TestKoStart(t *testing.T) {
	now := time.Date(2026, 6, 28, 3, 0, 0, 0, time.UTC) // 候29（λ≈96.5, 開始 λ=95）
	start := KoStart(now)

	if start.After(now) {
		t.Fatalf("KoStart が未来: start=%v now=%v", start, now)
	}
	// 1 候（約5.07日）以内に開始がある。
	if d := now.Sub(start); d <= 0 || d > 6*24*time.Hour {
		t.Errorf("now - KoStart = %v, want (0, ~6日]", d)
	}
	// 開始時刻の黄経は 5° の倍数（境界）。
	lambda := apparentSolarLongitudeDeg(start)
	frac := math.Mod(lambda, 5.0)
	if frac > 0.05 && frac < 4.95 {
		t.Errorf("KoStart の λ=%.4f が 5° 境界に乗っていない", lambda)
	}
	// 開始直後と直前で候番号が 1 つずれる（境界であることの確認）。
	if a, b := Number(start.Add(time.Minute)), Number(start.Add(-time.Minute)); a == b {
		t.Errorf("KoStart の前後で候が変わらない: %d", a)
	}
}

// TTLCutoff は N に対して単調に過去へ動き、N=1 では現在の候の開始に一致する。
func TestTTLCutoff(t *testing.T) {
	now := time.Date(2026, 6, 28, 3, 0, 0, 0, time.UTC)

	if c1, ks := TTLCutoff(now, 1), KoStart(now); !c1.Equal(ks) {
		t.Errorf("TTLCutoff(now,1)=%v, want KoStart=%v", c1, ks)
	}

	c1 := TTLCutoff(now, 1)
	c6 := TTLCutoff(now, 6)
	if !c6.Before(c1) {
		t.Errorf("TTLCutoff(6) は TTLCutoff(1) より過去のはず")
	}
	// 6候ぶんの遡りは概ね 5*平均候長。
	gotDays := c1.Sub(c6).Hours() / 24
	wantDays := 5 * meanKoDays
	if math.Abs(gotDays-wantDays) > 0.5 {
		t.Errorf("TTLCutoff(1)-TTLCutoff(6) = %.2f 日, want ~%.2f 日", gotDays, wantDays)
	}
}
