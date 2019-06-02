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

// AuthRegister export
func (rr *Router) AuthRegister(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		account := r.FormValue("account")
		password := r.FormValue("password")
		w.Header().Set("Content-Type", "application/json")
		response := rr.sf.Auth.Register(&account, &password)
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// AuthLogin export
func (rr *Router) AuthLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		account := r.FormValue("account")
		password := r.FormValue("password")
		w.Header().Set("Content-Type", "application/json")
		response := rr.sf.Auth.Login(&account, &password)
		js, _ := json.Marshal(response)
		w.Write(js)
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

// V1Novel export
func (rr *Router) V1Novel(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		novelID := r.URL.Query().Get("novel_id")
		if novelID != "" {
			w.Header().Set("Content-Type", "application/json")
			response := rr.sf.getNovelByID(&novelID)
			js, _ := json.Marshal(response)
			w.Write(js)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	case http.MethodPost:
		novelURL := r.FormValue("novel_url")
		if novelURL != "" {
			w.Header().Set("Content-Type", "application/json")
			response := rr.sf.getNovelByURL(&novelURL)
			js, _ := json.Marshal(response)
			w.Write(js)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// V1NovelChapter export
func (rr *Router) V1NovelChapter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		novelID := r.URL.Query().Get("novel_id")
		chapterIndex := r.URL.Query().Get("chapter_index")
		if novelID == "" || chapterIndex == "" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			response := rr.sf.getNovelChapter(&novelID, &chapterIndex)
			js, _ := json.Marshal(response)
			w.Write(js)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// V1Comics export
func (rr *Router) V1Comics(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		response := rr.sf.getComics()
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// V1Comic export
func (rr *Router) V1Comic(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		comicID := r.URL.Query().Get("comic_id")
		if comicID != "" {
			w.Header().Set("Content-Type", "application/json")
			response := rr.sf.getComicByID(&comicID)
			js, _ := json.Marshal(response)
			w.Write(js)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	case http.MethodPost:
		comicURL := r.FormValue("comic_url")
		if comicURL != "" {
			w.Header().Set("Content-Type", "application/json")
			response := rr.sf.getComicByURL(&comicURL)
			js, _ := json.Marshal(response)
			w.Write(js)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// V1ComicChapter export
func (rr *Router) V1ComicChapter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		comicID := r.URL.Query().Get("comic_id")
		chapterIndex := r.URL.Query().Get("chapter_index")
		if comicID == "" || chapterIndex == "" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			response := rr.sf.getComicChapter(&comicID, &chapterIndex)
			js, _ := json.Marshal(response)
			w.Write(js)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
