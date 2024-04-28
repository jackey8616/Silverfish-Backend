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
	authSer  *silverfish.Auth
	userSer  *silverfish.User
	novelSer *silverfish.Novel
	route    string
}

// NewBlueprintNovelv1 export
func NewBlueprintNovelv1(
	authSer *silverfish.Auth,
	userSer *silverfish.User,
	novelSer *silverfish.Novel,
) *BlueprintNovelv1 {
	bpn := new(BlueprintNovelv1)
	bpn.authSer = authSer
	bpn.userSer = userSer
	bpn.novelSer = novelSer
	bpn.route = "/novels"
	return bpn
}

// RouteRegister export
func (bpn *BlueprintNovelv1) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix(bpn.route).Subrouter()
	router.HandleFunc("", bpn.root).Methods("GET", "POST")
	router.HandleFunc("/", bpn.root).Methods("GET", "POST")
	router.HandleFunc("/{novelId}", bpn.novel).Methods("GET", "DELETE")
	router.HandleFunc("/{novelId}/chapter/{chapterIndex}", bpn.chapter).Methods("GET")
}

func (bpn *BlueprintNovelv1) root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	isAdmin := false
	sessionToken := r.Header.Get("Authorization")
	if sessionToken != "" {
		session, err := bpn.authSer.GetSession(&sessionToken)
		if err != nil {
			isAdmin = false
		} else if accountIsAdmin, _ := bpn.authSer.IsAdmin(session.GetAccount()); accountIsAdmin == true {
			isAdmin = true
		}
	}

	switch r.Method {
	case http.MethodGet:
		novelId := params["novelId"]
		if novelId != "" {
			result, err := bpn.novelSer.GetNovelById(&novelId)
			if (err != nil && err.Error() == "not found") ||
				(result != nil && !result.IsEnable && !isAdmin) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				response := entity.NewAPIResponse(result, err)
				js, _ := json.Marshal(response)
				w.Write(js)
			}
		} else {
			result, err := bpn.novelSer.GetNovels(isAdmin)
			response := entity.NewAPIResponse(result, err)
			js, _ := json.Marshal(response)
			w.Write(js)
		}
	case http.MethodPost:
		response := new(entity.APIResponse)
		if !isAdmin {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else {
			novelURL := r.FormValue("novel_url")
			if novelURL != "" {
				result, err := bpn.novelSer.AddNovelByURL(&novelURL)
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

func (bpn *BlueprintNovelv1) novel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	novelId := params["novelId"]
	if novelId == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	isAdmin := false
	sessionToken := r.Header.Get("Authorization")
	if sessionToken != "" {
		session, err := bpn.authSer.GetSession(&sessionToken)
		if err != nil {
			isAdmin = false
		} else if accountIsAdmin, _ := bpn.authSer.IsAdmin(session.GetAccount()); accountIsAdmin == true {
			isAdmin = true
		}
	}

	switch r.Method {
	case http.MethodGet:
		result, err := bpn.novelSer.GetNovelById(&novelId)
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
		} else if novelId != "" {
			err := bpn.novelSer.RemoveNovelById(&novelId)
			response = entity.NewAPIResponse(nil, err)
		} else {
			response = entity.NewAPIResponse(nil, errors.New("Field novelId should not be empty"))
		}
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (bpn *BlueprintNovelv1) chapter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	novelId := params["novelId"]
	chapterIndex := params["chapterIndex"]
	if novelId == "" || chapterIndex == "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	sessionToken := r.Header.Get("Authorization")
	session, _ := bpn.authSer.GetSession(&sessionToken)

	switch r.Method {
	case http.MethodGet:
		result, err := bpn.novelSer.GetNovelChapter(&novelId, &chapterIndex)
		response := entity.NewAPIResponse(result, err)
		if err == nil && session != nil {
			go bpn.userSer.UpdateBookmark("Novel", &novelId, session.GetAccount(), &chapterIndex)
		}
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
