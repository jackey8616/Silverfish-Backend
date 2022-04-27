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
	authSer  *silverfish.Auth
	userSer  *silverfish.User
	comicSer *silverfish.Comic
	route    string
}

// NewBlueprintComicv1 export
func NewBlueprintComicv1(
	authSer *silverfish.Auth,
	userSer *silverfish.User,
	comicSer *silverfish.Comic,
) *BlueprintComicv1 {
	bpc := new(BlueprintComicv1)
	bpc.authSer = authSer
	bpc.userSer = userSer
	bpc.comicSer = comicSer
	bpc.route = "/comics"
	return bpc
}

// RouteRegister export
func (bpc *BlueprintComicv1) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix(bpc.route).Subrouter()
	router.HandleFunc("", bpc.root).Methods("GET", "POST")
	router.HandleFunc("/", bpc.root).Methods("GET", "POST")
	router.HandleFunc("/{comicID}", bpc.comic).Methods("GET", "DELETE")
	router.HandleFunc("/{comicID}/chapter/{chapterIndex}", bpc.chapter).Methods("GET")
}

func (bpc *BlueprintComicv1) root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	isAdmin := false
	sessionToken := r.Header.Get("Authorization")
	if sessionToken != "" {
		session, err := bpc.authSer.GetSession(&sessionToken)
		if err != nil {
			isAdmin = false
		} else if accountIsAdmin, _ := bpc.authSer.IsAdmin(session.GetAccount()); accountIsAdmin == true {
			isAdmin = true
		}
	}

	switch r.Method {
	case http.MethodGet:
		comicID := params["comicID"]
		if comicID != "" {
			result, err := bpc.comicSer.GetComicByID(&comicID)
			if (err != nil && err.Error() == "not found") ||
				(result != nil && !result.IsEnable && !isAdmin) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				response := entity.NewAPIResponse(result, err)
				js, _ := json.Marshal(response)
				w.Write(js)
			}
		} else {
			result, err := bpc.comicSer.GetComics(isAdmin)
			response := entity.NewAPIResponse(result, err)
			js, _ := json.Marshal(response)
			w.Write(js)
		}
	case http.MethodPost:
		response := new(entity.APIResponse)
		if !isAdmin {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else {
			comicURL := r.FormValue("comic_url")
			if comicURL != "" {
				result, err := bpc.comicSer.AddComicByURL(&comicURL)
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

func (bpc *BlueprintComicv1) comic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	comicID := params["comicID"]
	if comicID == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	isAdmin := false
	sessionToken := r.Header.Get("Authorization")
	if sessionToken != "" {
		session, err := bpc.authSer.GetSession(&sessionToken)
		if err != nil {
			isAdmin = false
		} else if accountIsAdmin, _ := bpc.authSer.IsAdmin(session.GetAccount()); accountIsAdmin == true {
			isAdmin = true
		}
	}

	switch r.Method {
	case http.MethodGet:
		result, err := bpc.comicSer.GetComicByID(&comicID)
		if (err != nil && err.Error() == "not found") ||
			(result != nil && !result.IsEnable && !isAdmin) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			response := entity.NewAPIResponse(result, err)
			js, _ := json.Marshal(response)
			w.Write(js)
		}
	case http.MethodDelete:
		response := new(entity.APIResponse)
		if isAdmin == false {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else if comicID != "" {
			err := bpc.comicSer.RemoveComicByID(&comicID)
			response = entity.NewAPIResponse(nil, err)
		} else {
			response = entity.NewAPIResponse(nil, errors.New("Field comicID should not be empty"))
		}
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (bpc *BlueprintComicv1) chapter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	comicID := params["comicID"]
	chapterIndex := params["chapterIndex"]
	if comicID == "" || chapterIndex == "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	sessionToken := r.Header.Get("Authorization")
	session, _ := bpc.authSer.GetSession(&sessionToken)

	switch r.Method {
	case http.MethodGet:
		result, err := bpc.comicSer.GetComicChapter(&comicID, &chapterIndex)
		response := entity.NewAPIResponse(result, err)
		if err == nil && session != nil {
			go bpc.userSer.UpdateBookmark("Comic", &comicID, session.GetAccount(), &chapterIndex)
		}
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
