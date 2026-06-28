package ko

import (
	"testing"
	"time"
)

func TestSekkiOf(t *testing.T) {
	// 3 候で 1 節気。候1-3=立春(1)、候28-30=夏至(10)、候70-72=大寒(24)。
	cases := map[int]int{1: 1, 3: 1, 4: 2, 28: 10, 29: 10, 30: 10, 70: 24, 72: 24}
	for ko, want := range cases {
		if got := SekkiOf(ko); got != want {
			t.Errorf("SekkiOf(%d) = %d, want %d", ko, got, want)
		}
	}
}

func TestSekki(t *testing.T) {
	// 候29（菖蒲華）は夏至。
	s := Sekki(SekkiOf(29))
	if s.Number != 10 || s.Name != "夏至" || s.Kana != "げし" {
		t.Errorf("Sekki(夏至) = %+v, want {10 夏至 げし}", s)
	}
	// 端（立春・大寒）。
	if got := Sekki(1); got.Name != "立春" || got.Kana != "りっしゅん" {
		t.Errorf("Sekki(1) = %+v", got)
	}
	if got := Sekki(24); got.Name != "大寒" || got.Kana != "だいかん" {
		t.Errorf("Sekki(24) = %+v", got)
	}
}

func TestWafuMonthOf(t *testing.T) {
	if got := WafuMonthOf(time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)); got.Name != "水無月" || got.Kana != "みなづき" {
		t.Errorf("6 月 = %+v, want {水無月 みなづき}", got)
	}
	if got := WafuMonthOf(time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)); got.Name != "睦月" {
		t.Errorf("1 月 = %+v, want 睦月", got)
	}
	if got := WafuMonthOf(time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)); got.Name != "師走" {
		t.Errorf("12 月 = %+v, want 師走", got)
	}
}
