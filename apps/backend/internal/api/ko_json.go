package api

import "github.com/0muji4/Karin/apps/backend/internal/ko"

// koJSON は候メタの API 表現。今日の候と文箱の候別表示で共有する。
type koJSON struct {
	Number  int    `json:"number"`
	Name    string `json:"name"`
	Kana    string `json:"kana"`
	Meaning string `json:"meaning"`
	Sekki   int    `json:"sekki"`
	Season  string `json:"season"`
}

func toKoJSON(m ko.Meta) koJSON {
	return koJSON{
		Number:  m.Number,
		Name:    m.Name,
		Kana:    m.Kana,
		Meaning: m.Meaning,
		Sekki:   m.Sekki,
		Season:  m.Season,
	}
}
