package ko

import "time"

// SekkiMeta は二十四節気のメタ（名称・読み）。
type SekkiMeta struct {
	Number int // 1..24（立春起点）
	Name   string
	Kana   string
}

// sekkiTable は二十四節気（立春=1 起点）の名称と読み。年に依存しない固定の暦知識。
var sekkiTable = [24]struct{ name, kana string }{
	{"立春", "りっしゅん"}, {"雨水", "うすい"}, {"啓蟄", "けいちつ"}, {"春分", "しゅんぶん"},
	{"清明", "せいめい"}, {"穀雨", "こくう"},
	{"立夏", "りっか"}, {"小満", "しょうまん"}, {"芒種", "ぼうしゅ"}, {"夏至", "げし"},
	{"小暑", "しょうしょ"}, {"大暑", "たいしょ"},
	{"立秋", "りっしゅう"}, {"処暑", "しょしょ"}, {"白露", "はくろ"}, {"秋分", "しゅうぶん"},
	{"寒露", "かんろ"}, {"霜降", "そうこう"},
	{"立冬", "りっとう"}, {"小雪", "しょうせつ"}, {"大雪", "たいせつ"}, {"冬至", "とうじ"},
	{"小寒", "しょうかん"}, {"大寒", "だいかん"},
}

// SekkiOf は候番号（1..72）から二十四節気番号（1..24）を返す。3 候で 1 節気。
func SekkiOf(koNumber int) int {
	return (koNumber-1)/3 + 1
}

// Sekki は二十四節気番号（1..24）のメタを返す。
func Sekki(n int) SekkiMeta {
	t := sekkiTable[n-1]
	return SekkiMeta{Number: n, Name: t.name, Kana: t.kana}
}

// WafuMonth は和風月名（名称・読み）。
type WafuMonth struct {
	Name string
	Kana string
}

// wafuMonths は和風月名。現代の慣用に従いグレゴリオ月に対応づける（6 月＝水無月）。
var wafuMonths = [12]struct{ name, kana string }{
	{"睦月", "むつき"}, {"如月", "きさらぎ"}, {"弥生", "やよい"}, {"卯月", "うづき"},
	{"皐月", "さつき"}, {"水無月", "みなづき"}, {"文月", "ふみづき"}, {"葉月", "はづき"},
	{"長月", "ながつき"}, {"神無月", "かんなづき"}, {"霜月", "しもつき"}, {"師走", "しわす"},
}

// WafuMonthOf は時刻のグレゴリオ月に対応する和風月名を返す。
func WafuMonthOf(t time.Time) WafuMonth {
	m := wafuMonths[int(t.Month())-1]
	return WafuMonth{Name: m.name, Kana: m.kana}
}
