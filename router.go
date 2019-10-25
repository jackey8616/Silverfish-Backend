package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	silverfish "silverfish/silverfish"
	entity "silverfish/silverfish/entity"
	"strings"

	"github.com/pkg/errors"
)

// Router export
type Router struct {
	recaptchaPrivateKey *string
	sf                  *silverfish.Silverfish
	sessionUsecase      *silverfish.SessionUsecase
}

// NewRouter export
func NewRouter(recaptchaPrivateKey *string, sf *silverfish.Silverfish) *Router {
	rr := new(Router)
	rr.recaptchaPrivateKey = recaptchaPrivateKey
	rr.sf = sf
	rr.sessionUsecase = silverfish.NewSessionUsecase()
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

// Root export
func (rr *Router) Root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		js, _ := json.Marshal(map[string]bool{"Success": true})
		w.Write(js)
	}
}

// FetcherList export
func (rr *Router) FetcherList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sessionToken := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		session, err := rr.sessionUsecase.GetSession(&sessionToken)
		response := new(entity.APIResponse)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else if isAdmin, _ := rr.sf.Auth.IsAdmin(session.GetAccount()); isAdmin == false {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else {
			fetcherLists := rr.sf.GetLists()
			response = entity.NewAPIResponse(map[string]interface{}{
				"fetchers": fetcherLists,
			}, nil)
		}
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
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
			user, err := rr.sf.Auth.Register(false, &account, &password)
			response = entity.NewAPIResponse(map[string]interface{}{
				"session": *rr.sessionUsecase.InsertSession(user, false),
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
						"session": *rr.sessionUsecase.InsertSession(user, keepLogin == "true"),
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

// AuthIsAdmin export
func (rr *Router) AuthIsAdmin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sessionToken := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		session, err := rr.sessionUsecase.GetSession(&sessionToken)
		response := new(entity.APIResponse)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else {
			isAdmin, _ := rr.sf.Auth.IsAdmin(session.GetAccount())
			response = entity.NewAPIResponse(map[string]interface{}{
				"isAdmin": isAdmin,
			}, nil)
		}
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

		result, err := rr.sf.GetNovels()
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

			result, err := rr.sf.GetNovelByID(&novelID)
			response := entity.NewAPIResponse(result, err)
			js, _ := json.Marshal(response)
			w.Write(js)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	case http.MethodPost:
		sessionToken := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		session, err := rr.sessionUsecase.GetSession(&sessionToken)
		response := new(entity.APIResponse)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else if isAdmin, _ := rr.sf.Auth.IsAdmin(session.GetAccount()); isAdmin == false {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else {
			novelURL := r.FormValue("novel_url")
			if novelURL != "" {
				w.Header().Set("Content-Type", "application/json")

				result, err := rr.sf.GetNovelByURL(&novelURL)
				response = entity.NewAPIResponse(result, err)
			} else {
				response = entity.NewAPIResponse(nil, errors.New("Field novel_url should not be empty"))
			}
		}
		js, _ := json.Marshal(response)
		w.Write(js)
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

			result, err := rr.sf.GetNovelChapter(&novelID, &chapterIndex)
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

		result, err := rr.sf.GetComics()
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

			result, err := rr.sf.GetComicByID(&comicID)
			response := entity.NewAPIResponse(result, err)
			js, _ := json.Marshal(response)
			w.Write(js)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	case http.MethodPost:
		sessionToken := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		session, err := rr.sessionUsecase.GetSession(&sessionToken)
		response := new(entity.APIResponse)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else if isAdmin, _ := rr.sf.Auth.IsAdmin(session.GetAccount()); isAdmin == false {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else {
			comicURL := r.FormValue("comic_url")
			if comicURL != "" {
				w.Header().Set("Content-Type", "application/json")

				result, err := rr.sf.GetComicByURL(&comicURL)
				response = entity.NewAPIResponse(result, err)
			} else {
				response = entity.NewAPIResponse(nil, errors.New("Field comic_url should not be empty"))
			}
		}
		js, _ := json.Marshal(response)
		w.Write(js)
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
			result, err := rr.sf.GetComicChapter(&comicID, &chapterIndex)
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
