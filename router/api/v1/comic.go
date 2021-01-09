package v1

import (
	"encoding/json"
	"net/http"

	silverfish "silverfish/silverfish"
	entity "silverfish/silverfish/entity"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// BlueprintComicv1 export
type BlueprintComicv1 struct {
	silverfish     *silverfish.Silverfish
	auth           *silverfish.Auth
	sessionUsecase *silverfish.SessionUsecase
	route          string
}

// NewBlueprintComicv1 export
func NewBlueprintComicv1(silverfish *silverfish.Silverfish, sessionUsecase *silverfish.SessionUsecase) *BlueprintComicv1 {
	bpc := new(BlueprintComicv1)
	bpc.silverfish = silverfish
	bpc.auth = silverfish.Auth
	bpc.sessionUsecase = sessionUsecase
	bpc.route = "/comic"
	return bpc
}

// RouteRegister export
func (bpc *BlueprintComicv1) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix(bpc.route).Subrouter()
	router.HandleFunc("", bpc.comic).Methods("GET", "POST")
	router.HandleFunc("/", bpc.comic).Methods("GET", "POST")
	router.HandleFunc("/{comicID}", bpc.deleteComic).Methods("DELETE")
	router.HandleFunc("/chapter", bpc.chapter).Methods("GET")

	sRouter := parentRouter.PathPrefix(bpc.route + "s").Subrouter()
	sRouter.HandleFunc("", bpc.listComic).Methods("GET")
	sRouter.HandleFunc("/", bpc.listComic).Methods("GET")
	sRouter.HandleFunc("/{comicID}", bpc.deleteComic).Methods("DELETE")
	sRouter.HandleFunc("/chapter", bpc.chapter).Methods("GET")
}

func (bpc *BlueprintComicv1) comic(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		comicID := r.URL.Query().Get("comic_id")
		if comicID != "" {
			w.Header().Set("Content-Type", "application/json")

			result, err := bpc.silverfish.GetComicByID(&comicID)
			response := entity.NewAPIResponse(result, err)
			js, _ := json.Marshal(response)
			w.Write(js)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	case http.MethodPost:
		sessionToken := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		session, err := bpc.sessionUsecase.GetSession(&sessionToken)
		response := new(entity.APIResponse)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else if isAdmin, _ := bpc.auth.IsAdmin(session.GetAccount()); isAdmin == false {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else {
			comicURL := r.FormValue("comic_url")
			if comicURL != "" {
				w.Header().Set("Content-Type", "application/json")

				result, err := bpc.silverfish.AddComicByURL(&comicURL)
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

func (bpc *BlueprintComicv1) deleteComic(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		sessionToken := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		session, err := bpc.sessionUsecase.GetSession(&sessionToken)
		response := new(entity.APIResponse)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else if isAdmin, _ := bpc.auth.IsAdmin(session.GetAccount()); isAdmin == false {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else {
			params := mux.Vars(r)
			comicID := params["comicID"]

			if comicID != "" {
				err := bpc.silverfish.RemoveComicByID(&comicID)
				response = entity.NewAPIResponse(nil, err)
			} else {
				response = entity.NewAPIResponse(nil, errors.New("Field comic_id should not be empty"))
			}
		}
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (bpc *BlueprintComicv1) listComic(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		shouldFetchDisable := false
		sessionToken := r.Header.Get("Authorization")
		if sessionToken != "" {
			session, _ := bpc.sessionUsecase.GetSession(&sessionToken)
			if session == nil {
				shouldFetchDisable = false
			} else if isAdmin, _ := bpc.auth.IsAdmin(session.GetAccount()); isAdmin == true {
				shouldFetchDisable = true
			}
		}

		result, err := bpc.silverfish.GetComics(shouldFetchDisable)
		response := entity.NewAPIResponse(result, err)
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (bpc *BlueprintComicv1) chapter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sessionToken := r.Header.Get("Authorization")
		session, _ := bpc.sessionUsecase.GetSession(&sessionToken)

		comicID := r.URL.Query().Get("comic_id")
		chapterIndex := r.URL.Query().Get("chapter_index")
		if comicID == "" || chapterIndex == "" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			result, err := bpc.silverfish.GetComicChapter(&comicID, &chapterIndex)
			response := entity.NewAPIResponse(result, err)
			if err == nil && session != nil {
				go bpc.auth.UpdateBookmark("Comic", &comicID, session.GetAccount(), &chapterIndex)
			}
			js, _ := json.Marshal(response)
			w.Write(js)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
