// Package ko は時刻から七十二候（1..72）を太陽の見かけの黄経で算出する。
//
// 七十二候は二十四節気をさらに 3 分割した、約 5 日ごとの暦。候の境界は太陽黄経が
// 5° 進むごとに移り、候1（東風解凍, 立春）は黄経 315° から始まる。境界の日付は年ごとに
// 数日ずれる天文現象なので、年別の日付表を持たず計算で解決する（設計判断: データではなく機構）。
package ko

import (
	"math"
	"time"
)

// Min と Max は候番号の範囲。
const (
	Min = 1
	Max = 72
)

// risshunIndex は立春の 5° バンド番号（315° / 5°）。候1 の起点。
const risshunIndex = 63

// Number は時刻 t の七十二候（1..72）を返す。
// 黄経は瞬時の値なので、結果はタイムゾーンに依存しない（t の絶対時刻だけで決まる）。
func Number(t time.Time) int {
	return numberFromLongitude(apparentSolarLongitudeDeg(t))
}

// numberFromLongitude は黄経（度）を七十二候（1..72）に写す。
// 5° バンドで区切り、候1 は λ=315°（立春）から始まる。天文計算と分けて単体検証できる。
func numberFromLongitude(lambda float64) int {
	idx := int(math.Floor(norm360(lambda) / 5.0)) // 0..71（黄経の 5° バンド）
	return ((idx-risshunIndex)%72+72)%72 + 1
}

// apparentSolarLongitudeDeg は太陽の見かけの黄経を度で返す（[0,360)）。
// Meeus『天文計算』25 章の低精度版。誤差は約 0.01° で、候境界（5° 幅・約 5 日）の
// 判定には十分すぎる精度。
func apparentSolarLongitudeDeg(t time.Time) float64 {
	jd := julianDay(t)
	tc := (jd - 2451545.0) / 36525.0 // J2000.0 からのユリウス世紀

	// 太陽の幾何平均黄経と平均近点角（度）。
	l0 := 280.46646 + 36000.76983*tc + 0.0003032*tc*tc
	m := 357.52911 + 35999.05029*tc - 0.0001537*tc*tc
	mr := rad(m)

	// 中心差（the equation of center）。
	c := (1.914602-0.004817*tc-0.000014*tc*tc)*math.Sin(mr) +
		(0.019993-0.000101*tc)*math.Sin(2*mr) +
		0.000289*math.Sin(3*mr)

	trueLong := l0 + c

	// 章動・光行差を補正して見かけの黄経へ。
	omega := 125.04 - 1934.136*tc
	apparent := trueLong - 0.00569 - 0.00478*math.Sin(rad(omega))
	return norm360(apparent)
}

// julianDay は時刻 t のユリウス日を返す（Unix 紀元 1970-01-01T00:00:00Z = JD 2440587.5）。
func julianDay(t time.Time) float64 {
	return float64(t.UTC().UnixNano())/1e9/86400.0 + 2440587.5
}

func rad(deg float64) float64 { return deg * math.Pi / 180.0 }

func norm360(d float64) float64 {
	d = math.Mod(d, 360.0)
	if d < 0 {
		d += 360.0
	}
	return d
}
