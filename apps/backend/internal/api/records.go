package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/ko"
	"github.com/0muji4/Karin/apps/backend/internal/record"
)

type createRecordRequest struct {
	Body string `json:"body"`
	// KoWritten は省略可。省略時は今日の候を使う。
	KoWritten *int `json:"ko_written"`
}

type recordJSON struct {
	ID        string    `json:"id"`
	KoWritten int       `json:"ko_written"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func toRecordJSON(r record.Record) recordJSON {
	return recordJSON{
		ID:        r.ID.String(),
		KoWritten: r.KoWritten,
		Body:      r.Body,
		CreatedAt: r.CreatedAt,
	}
}

// handleCreateRecord は本人の文箱に短冊を保存する。候の指定が無ければ今日の候を使う。
func (s *Server) handleCreateRecord(w http.ResponseWriter, r *http.Request) {
	p, ok := principalFrom(r.Context())
	if !ok {
		writeError(w, s.logger, http.StatusUnauthorized, "unauthorized", "認証が必要")
		return
	}

	var req createRecordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, s.logger, http.StatusBadRequest, "invalid_json", "リクエストの形式が不正")
		return
	}

	koNum := ko.Number(time.Now())
	if req.KoWritten != nil {
		koNum = *req.KoWritten
	}

	rec, err := s.records.Create(r.Context(), p.UserID, req.Body, koNum)
	if errors.Is(err, record.ErrInvalid) {
		writeError(w, s.logger, http.StatusBadRequest, "invalid_record", err.Error())
		return
	}
	if err != nil {
		s.logger.Error("記録の保存に失敗", "error", err)
		writeError(w, s.logger, http.StatusInternalServerError, "internal", "記録を保存できなかった")
		return
	}
	writeJSON(w, s.logger, http.StatusCreated, toRecordJSON(rec))
}

type boxGroup struct {
	Ko      koJSON       `json:"ko"`
	Records []recordJSON `json:"records"`
}

type boxResponse struct {
	Groups []boxGroup `json:"groups"`
}

// handleListBox は本人の文箱を候別にまとめて返す（候昇順、各候内は新しい順）。
func (s *Server) handleListBox(w http.ResponseWriter, r *http.Request) {
	p, ok := principalFrom(r.Context())
	if !ok {
		writeError(w, s.logger, http.StatusUnauthorized, "unauthorized", "認証が必要")
		return
	}

	recs, err := s.records.ListByOwner(r.Context(), p.UserID)
	if err != nil {
		s.logger.Error("文箱の取得に失敗", "error", err)
		writeError(w, s.logger, http.StatusInternalServerError, "internal", "文箱を取得できなかった")
		return
	}

	metaByKo, err := s.koMetaMap(r)
	if err != nil {
		s.logger.Error("候メタの取得に失敗", "error", err)
		writeError(w, s.logger, http.StatusInternalServerError, "internal", "文箱を取得できなかった")
		return
	}

	// recs は候昇順に並んでいるので、候が変わる境目でグループを切る。
	groups := make([]boxGroup, 0)
	for i := 0; i < len(recs); {
		koNum := recs[i].KoWritten
		g := boxGroup{Ko: koJSONFor(koNum, metaByKo), Records: []recordJSON{}}
		for i < len(recs) && recs[i].KoWritten == koNum {
			g.Records = append(g.Records, toRecordJSON(recs[i]))
			i++
		}
		groups = append(groups, g)
	}
	writeJSON(w, s.logger, http.StatusOK, boxResponse{Groups: groups})
}

// koMetaMap は全候のメタを番号引きの map で返す。
func (s *Server) koMetaMap(r *http.Request) (map[int]sqlcdb.KoReference, error) {
	metas, err := sqlcdb.New(s.pool).ListKoReference(r.Context())
	if err != nil {
		return nil, err
	}
	m := make(map[int]sqlcdb.KoReference, len(metas))
	for _, meta := range metas {
		m[int(meta.Ko)] = meta
	}
	return m, nil
}

// koJSONFor はメタがあれば埋め、無ければ番号だけ返す。
func koJSONFor(koNum int, metaByKo map[int]sqlcdb.KoReference) koJSON {
	if m, ok := metaByKo[koNum]; ok {
		return toKoJSON(m)
	}
	return koJSON{Number: koNum}
}
