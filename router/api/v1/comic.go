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
	auth  *silverfish.Auth
	user  *silverfish.User
	comic *silverfish.Comic
	route string
}

// NewBlueprintComicv1 export
func NewBlueprintComicv1(
	auth *silverfish.Auth,
	user *silverfish.User,
	comic *silverfish.Comic,
) *BlueprintComicv1 {
	bpc := new(BlueprintComicv1)
	bpc.auth = auth
	bpc.user = user
	bpc.comic = comic
	bpc.auth = auth
	bpc.route = "/comic"
	return bpc
}

// RouteRegister export
func (bpc *BlueprintComicv1) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix(bpc.route).Subrouter()
	router.HandleFunc("", bpc.root).Methods("GET", "POST")
	router.HandleFunc("/", bpc.root).Methods("GET", "POST")
	router.HandleFunc("/{comicID}", bpc.deleteComic).Methods("DELETE")
	router.HandleFunc("/chapter", bpc.chapter).Methods("GET")

	sRouter := parentRouter.PathPrefix(bpc.route + "s").Subrouter()
	sRouter.HandleFunc("", bpc.listComic).Methods("GET")
	sRouter.HandleFunc("/", bpc.listComic).Methods("GET")
	sRouter.HandleFunc("/{comicID}", bpc.deleteComic).Methods("DELETE")
	sRouter.HandleFunc("/chapter", bpc.chapter).Methods("GET")
}

func (bpc *BlueprintComicv1) root(w http.ResponseWriter, r *http.Request) {
	isAdmin := false
	sessionToken := r.Header.Get("Authorization")
	if sessionToken != "" {
		session, err := bpc.auth.GetSession(&sessionToken)
		if err != nil {
			isAdmin = false
		} else if accountIsAdmin, _ := bpc.auth.IsAdmin(session.GetAccount()); accountIsAdmin == true {
			isAdmin = true
		}
	}

	switch r.Method {
	case http.MethodGet:
		comicID := r.URL.Query().Get("comic_id")
		if comicID != "" {
			w.Header().Set("Content-Type", "application/json")

			response := new(entity.APIResponse)
			result, err := bpc.comic.GetComicByID(&comicID)
			if (err != nil && err.Error() == "not found") ||
				(result != nil && !result.IsEnable && !isAdmin) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				response = entity.NewAPIResponse(result, err)
			}
			js, _ := json.Marshal(response)
			w.Write(js)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	case http.MethodPost:
		w.Header().Set("Content-Type", "application/json")

		response := new(entity.APIResponse)
		if !isAdmin {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else {
			comicURL := r.FormValue("comic_url")
			if comicURL != "" {
				w.Header().Set("Content-Type", "application/json")

				result, err := bpc.comic.AddComicByURL(&comicURL)
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

		session, err := bpc.auth.GetSession(&sessionToken)
		response := new(entity.APIResponse)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else if isAdmin, _ := bpc.auth.IsAdmin(session.GetAccount()); isAdmin == false {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else {
			params := mux.Vars(r)
			comicID := params["comicID"]

			if comicID != "" {
				err := bpc.comic.RemoveComicByID(&comicID)
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
			session, _ := bpc.auth.GetSession(&sessionToken)
			if session == nil {
				shouldFetchDisable = false
			} else if isAdmin, _ := bpc.auth.IsAdmin(session.GetAccount()); isAdmin == true {
				shouldFetchDisable = true
			}
		}

		result, err := bpc.comic.GetComics(shouldFetchDisable)
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
		session, _ := bpc.auth.GetSession(&sessionToken)

		comicID := r.URL.Query().Get("comic_id")
		chapterIndex := r.URL.Query().Get("chapter_index")
		if comicID == "" || chapterIndex == "" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			result, err := bpc.comic.GetComicChapter(&comicID, &chapterIndex)
			response := entity.NewAPIResponse(result, err)
			if err == nil && session != nil {
				go bpc.user.UpdateBookmark("Comic", &comicID, session.GetAccount(), &chapterIndex)
			}
			js, _ := json.Marshal(response)
			w.Write(js)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
