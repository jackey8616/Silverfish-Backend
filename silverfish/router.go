package silverfish

import (
	"encoding/json"
	"fmt"
	"net/http"
	entity "silverfish/silverfish/entity"
	"strings"

	"github.com/pkg/errors"
)

// Router export
type Router struct {
	recaptchaPrivateKey *string
	sf                  *Silverfish
}

// NewRouter export
func NewRouter(recaptchaPrivateKey *string, sf *Silverfish) *Router {
	rr := new(Router)
	rr.recaptchaPrivateKey = recaptchaPrivateKey
	rr.sf = sf
	return rr
}

func (rr *Router) verifyRecaptcha(token *string) (bool, error) {
	url := fmt.Sprintf(
		"https://www.google.com/recaptcha/api/siteverify?secret=%s&response=%s",
		*rr.recaptchaPrivateKey,
		*token)
	res, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	result := new(entity.RecaptchaResponse)
	json.NewDecoder(res.Body).Decode(&result)
	return result.Success, errors.New(strings.Join(result.ErrorCodes, " , "))
}

// AuthRegister export
func (rr *Router) AuthRegister(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		account := r.FormValue("account")
		password := r.FormValue("password")
		recaptchaToken := r.FormValue("recaptchaToken")
		w.Header().Set("Content-Type", "application/json")
		response := new(entity.APIResponse)
		if recaptchaToken == "" {
			response = &entity.APIResponse{
				Fail: true,
				Data: map[string]string{"reason": "Missing recaptcha token."},
			}
		} else if res, err := rr.verifyRecaptcha(&recaptchaToken); res == false {
			fmt.Println(err)
			response = &entity.APIResponse{
				Fail: true,
				Data: map[string]string{"reason": "Recaptcha verify failed."},
			}
		} else {
			response = rr.sf.Auth.Register(&account, &password)
		}
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
		recaptchaToken := r.FormValue("recaptchaToken")
		w.Header().Set("Content-Type", "application/json")
		response := new(entity.APIResponse)
		if recaptchaToken == "" {
			response = &entity.APIResponse{
				Fail: true,
				Data: map[string]string{"reason": "Missing recaptcha token."},
			}
		} else if res, err := rr.verifyRecaptcha(&recaptchaToken); res == false {
			fmt.Println(err)
			response = &entity.APIResponse{
				Fail: true,
				Data: map[string]string{"reason": "Recaptcha verify failed."},
			}
		} else {
			response = rr.sf.Auth.Login(&account, &password)
		}
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
		account := r.Header.Get("Reader")
		novelID := r.URL.Query().Get("novel_id")
		chapterIndex := r.URL.Query().Get("chapter_index")
		if novelID == "" || chapterIndex == "" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			response := rr.sf.getNovelChapter(&novelID, &chapterIndex)
			if account != "" && account != "guest" {
				go rr.sf.Auth.UpdateBookmark("Novel", &novelID, &account, &chapterIndex)
			}
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
		account := r.Header.Get("Reader")
		comicID := r.URL.Query().Get("comic_id")
		chapterIndex := r.URL.Query().Get("chapter_index")
		if comicID == "" || chapterIndex == "" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			response := rr.sf.getComicChapter(&comicID, &chapterIndex)
			if account != "" && account != "guest" {
				go rr.sf.Auth.UpdateBookmark("Comic", &comicID, &account, &chapterIndex)
			}
			js, _ := json.Marshal(response)
			w.Write(js)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
