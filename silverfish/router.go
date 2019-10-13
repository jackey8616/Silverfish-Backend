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
	sessionUsecase      *SessionUsecase
}

// NewRouter export
func NewRouter(recaptchaPrivateKey *string, sf *Silverfish) *Router {
	rr := new(Router)
	rr.recaptchaPrivateKey = recaptchaPrivateKey
	rr.sf = sf
	rr.sessionUsecase = NewSessionUsecase()
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

// AuthStatus export
func (rr *Router) AuthStatus(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sessionToken := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		result := rr.sessionUsecase.IsTokenValid(&sessionToken)
		response := &entity.APIResponse{
			Success: true,
			Data:    result,
		}
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
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
			response = entity.NewAPIResponse(nil, errors.New("Missing recaptcha token"))
		} else if res, err := rr.verifyRecaptcha(&recaptchaToken); res == false {
			fmt.Println(err)
			response = entity.NewAPIResponse(nil, errors.New("Recaptcha verify failed"))
		} else {
			user, err := rr.sf.Auth.Register(&account, &password)
			response = entity.NewAPIResponse(map[string]interface{}{
				"session": *rr.sessionUsecase.InsertSession(&account, false),
				"user":    user,
			}, err)
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
		keepLogin := r.FormValue("keepLogin")
		account := r.FormValue("account")
		password := r.FormValue("password")
		recaptchaToken := r.FormValue("recaptchaToken")
		w.Header().Set("Content-Type", "application/json")

		response := new(entity.APIResponse)
		if recaptchaToken == "" {
			response = entity.NewAPIResponse(nil, errors.New("Missing recaptcha token"))
		} else if res, err := rr.verifyRecaptcha(&recaptchaToken); res == false {
			fmt.Println(err)
			response = entity.NewAPIResponse(nil, errors.New("Recaptcha verify failed"))
		} else {
			user, err := rr.sf.Auth.Login(&account, &password)
			if err != nil {
				response = entity.NewAPIResponse(nil, err)
			} else {
				response = entity.NewAPIResponse(
					map[string]interface{}{
						"session": *rr.sessionUsecase.InsertSession(&account, keepLogin == "true"),
						"user":    user,
					}, nil)
			}
		}
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// AuthLogout export
func (rr *Router) AuthLogout(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sessionToken := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		rr.sessionUsecase.KillSession(&sessionToken)
		response := entity.NewAPIResponse(nil, nil)
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// UserBookmark export
func (rr *Router) UserBookmark(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sessionToken := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		session, err := rr.sessionUsecase.GetSession(&sessionToken)
		response := new(entity.APIResponse)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else {
			result, err := rr.sf.Auth.GetUserBookmark(session.GetAccount())
			response = entity.NewAPIResponse(result, err)
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

		result, err := rr.sf.getNovels()
		response := entity.NewAPIResponse(result, err)
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

			result, err := rr.sf.getNovelByID(&novelID)
			response := entity.NewAPIResponse(result, err)
			js, _ := json.Marshal(response)
			w.Write(js)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	case http.MethodPost:
		novelURL := r.FormValue("novel_url")
		if novelURL != "" {
			w.Header().Set("Content-Type", "application/json")

			result, err := rr.sf.getNovelByURL(&novelURL)
			response := entity.NewAPIResponse(result, err)
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
		sessionToken := r.Header.Get("Authorization")
		session, _ := rr.sessionUsecase.GetSession(&sessionToken)

		novelID := r.URL.Query().Get("novel_id")
		chapterIndex := r.URL.Query().Get("chapter_index")
		if novelID == "" || chapterIndex == "" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")

			result, err := rr.sf.getNovelChapter(&novelID, &chapterIndex)
			response := entity.NewAPIResponse(result, err)
			if err == nil && session != nil {
				go rr.sf.Auth.UpdateBookmark("Novel", &novelID, session.GetAccount(), &chapterIndex)
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

		result, err := rr.sf.getComics()
		response := entity.NewAPIResponse(result, err)
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

			result, err := rr.sf.getComicByID(&comicID)
			response := entity.NewAPIResponse(result, err)
			js, _ := json.Marshal(response)
			w.Write(js)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	case http.MethodPost:
		comicURL := r.FormValue("comic_url")
		if comicURL != "" {
			w.Header().Set("Content-Type", "application/json")

			result, err := rr.sf.getComicByURL(&comicURL)
			response := entity.NewAPIResponse(result, err)
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
		sessionToken := r.Header.Get("Authorization")
		session, _ := rr.sessionUsecase.GetSession(&sessionToken)

		comicID := r.URL.Query().Get("comic_id")
		chapterIndex := r.URL.Query().Get("chapter_index")
		if comicID == "" || chapterIndex == "" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			result, err := rr.sf.getComicChapter(&comicID, &chapterIndex)
			response := entity.NewAPIResponse(result, err)
			if err == nil && session != nil {
				go rr.sf.Auth.UpdateBookmark("Comic", &comicID, session.GetAccount(), &chapterIndex)
			}
			js, _ := json.Marshal(response)
			w.Write(js)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
