package ko

import (
	"math"
	"time"
)

// meanKoDays は候の平均長（熱帯年 / 72 候 ≈ 5.07 日）。候TTL の遡り計算に使う。
const meanKoDays = 365.2422 / 72.0

// KoStart は時刻 t が属する候の開始時刻（直近で太陽黄経が 5° 境界を跨いだ瞬間）を返す。
func KoStart(t time.Time) time.Time {
	lambda := apparentSolarLongitudeDeg(t)
	bandStart := math.Floor(lambda/5.0) * 5.0 // この候の開始黄経（5° の倍数, [0,360)）
	// 直近 8 日で λ は約 8° 進むので、現在の候の開始境界を必ず内側に挟める。
	lo := t.Add(-8 * 24 * time.Hour)
	return solveCrossing(bandStart, lo, t)
}

// TTLCutoff は「N候のあいだ未配信」を判定する境界時刻を返す。
// この時刻より前にプールされた短冊は、N候以上滞留したものとみなして expired にする。
// 現在の候の開始（厳密）から、平均候長で (N-1) 候ぶん遡る。鮮度は緩い変数のため平均長で十分。
func TTLCutoff(now time.Time, nKo int) time.Time {
	start := KoStart(now)
	back := time.Duration(float64(nKo-1) * meanKoDays * 24 * float64(time.Hour))
	return start.Add(-back)
}

// solveCrossing は [lo,hi] で λ(t) が target（5° 境界）を跨ぐ瞬間を二分法で求める。
// λ は区間内で単調増加とみなせる。0/360 の折り返しは mod で吸収する。
// 前提: f(lo) < 0（境界より手前）, f(hi) >= 0（境界以降）。
func solveCrossing(target float64, lo, hi time.Time) time.Time {
	f := func(t time.Time) float64 {
		// (λ - target) を (-180,180] に正規化。境界の前後で符号が変わる。
		return math.Mod(apparentSolarLongitudeDeg(t)-target+540, 360) - 180
	}
	negLo := f(lo) < 0
	for i := 0; i < 48; i++ {
		mid := lo.Add(hi.Sub(lo) / 2)
		if (f(mid) < 0) == negLo {
			lo = mid
		} else {
			hi = mid
		}
	}
	return lo.Add(hi.Sub(lo) / 2)
}
