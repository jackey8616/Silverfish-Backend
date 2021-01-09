package v1

import (
	"encoding/json"
	"net/http"

	silverfish "silverfish/silverfish"
	entity "silverfish/silverfish/entity"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// BlueprintNovelv1 export
type BlueprintNovelv1 struct {
	silverfish     *silverfish.Silverfish
	auth           *silverfish.Auth
	sessionUsecase *silverfish.SessionUsecase
	route          string
}

// NewBlueprintNovelv1 export
func NewBlueprintNovelv1(silverfish *silverfish.Silverfish, sessionUsecase *silverfish.SessionUsecase) *BlueprintNovelv1 {
	bpn := new(BlueprintNovelv1)
	bpn.silverfish = silverfish
	bpn.auth = silverfish.Auth
	bpn.sessionUsecase = sessionUsecase
	bpn.route = "/novel"
	return bpn
}

// RouteRegister export
func (bpn *BlueprintNovelv1) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix(bpn.route).Subrouter()
	router.HandleFunc("", bpn.novel).Methods("GET", "POST")
	router.HandleFunc("/", bpn.novel).Methods("GET", "POST")
	router.HandleFunc("/{novelID}", bpn.deleteNovel).Methods("DELETE")
	/* TODO: route should be /api/v1/novel/chapter
	This change will need to update Frontend's api calling. */
	parentRouter.HandleFunc("/chapter", bpn.chapter).Methods("GET")
	router.HandleFunc("/chapter", bpn.chapter).Methods("GET")

	sRouter := parentRouter.PathPrefix(bpn.route + "s").Subrouter()
	sRouter.HandleFunc("", bpn.listNovel).Methods("GET")
	sRouter.HandleFunc("/", bpn.listNovel).Methods("GET")
	sRouter.HandleFunc("/{novelID}", bpn.deleteNovel).Methods("DELETE")
	sRouter.HandleFunc("/chapter", bpn.chapter).Methods("GET")
}

func (bpn *BlueprintNovelv1) novel(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		novelID := r.URL.Query().Get("novel_id")
		if novelID != "" {
			w.Header().Set("Content-Type", "application/json")

			result, err := bpn.silverfish.GetNovelByID(&novelID)
			response := entity.NewAPIResponse(result, err)
			js, _ := json.Marshal(response)
			w.Write(js)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	case http.MethodPost:
		sessionToken := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		session, err := bpn.sessionUsecase.GetSession(&sessionToken)
		response := new(entity.APIResponse)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else if isAdmin, _ := bpn.auth.IsAdmin(session.GetAccount()); isAdmin == false {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else {
			novelURL := r.FormValue("novel_url")
			if novelURL != "" {
				w.Header().Set("Content-Type", "application/json")

				result, err := bpn.silverfish.AddNovelByURL(&novelURL)
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

func (bpn *BlueprintNovelv1) deleteNovel(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		sessionToken := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		session, err := bpn.sessionUsecase.GetSession(&sessionToken)
		response := new(entity.APIResponse)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else if isAdmin, _ := bpn.auth.IsAdmin(session.GetAccount()); isAdmin == false {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else {
			params := mux.Vars(r)
			novelID := params["novelID"]

			if novelID != "" {
				err := bpn.silverfish.RemoveNovelByID(&novelID)
				response = entity.NewAPIResponse(nil, err)
			} else {
				response = entity.NewAPIResponse(nil, errors.New("Field novel_id should not be empty"))
			}
		}
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (bpn *BlueprintNovelv1) listNovel(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		shouldFetchDisable := false
		sessionToken := r.Header.Get("Authorization")
		if sessionToken != "" {
			session, _ := bpn.sessionUsecase.GetSession(&sessionToken)
			if session == nil {
				shouldFetchDisable = false
			} else if isAdmin, _ := bpn.auth.IsAdmin(session.GetAccount()); isAdmin == true {
				shouldFetchDisable = true
			}
		}

		result, err := bpn.silverfish.GetNovels(shouldFetchDisable)
		response := entity.NewAPIResponse(result, err)
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (bpn *BlueprintNovelv1) chapter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sessionToken := r.Header.Get("Authorization")
		session, _ := bpn.sessionUsecase.GetSession(&sessionToken)

		novelID := r.URL.Query().Get("novel_id")
		chapterIndex := r.URL.Query().Get("chapter_index")
		if novelID == "" || chapterIndex == "" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")

			result, err := bpn.silverfish.GetNovelChapter(&novelID, &chapterIndex)
			response := entity.NewAPIResponse(result, err)
			if err == nil && session != nil {
				go bpn.auth.UpdateBookmark("Novel", &novelID, session.GetAccount(), &chapterIndex)
			}
			js, _ := json.Marshal(response)
			w.Write(js)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
