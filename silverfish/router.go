package silverfish

import (
	"encoding/json"
	"net/http"
)

// Router export
type Router struct {
	sf *Silverfish
}

// NewRouter export
func NewRouter(sf *Silverfish) *Router {
	rr := new(Router)
	rr.sf = sf
	return rr
}

// V1Novel export
func (rr *Router) V1Novel(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		novelID := r.URL.Query().Get("novel_id")
		if novelID == "" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			response := rr.sf.getNovel(&novelID)
			js, _ := json.Marshal(response)
			w.Write(js)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// V1Novels export
func (rr *Router) V1Novels(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		response := rr.sf.getNovels()
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// V1Chapter export
func (rr *Router) V1Chapter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		novelID := r.URL.Query().Get("novel_id")
		if novelID == "" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			response := rr.sf.getChapter(&novelID)
			js, _ := json.Marshal(response)
			w.Write(js)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// V1ChapterNew export
func (rr *Router) V1ChapterNew(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		novelID := r.URL.Query().Get("novel_id")
		chapterIndex := r.URL.Query().Get("chapter_index")
		if novelID == "" || chapterIndex == "" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			response := rr.sf.getChapterNew(&novelID, &chapterIndex)
			js, _ := json.Marshal(response)
			w.Write(js)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
