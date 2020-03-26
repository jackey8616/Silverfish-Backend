package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	interf "silverfish/router/interface"
	silverfish "silverfish/silverfish"
	entity "silverfish/silverfish/entity"

	"github.com/gorilla/mux"
)

// BlueprintUser export
type BlueprintUser struct {
	silverfish     *silverfish.Silverfish
	auth           *silverfish.Auth
	sessionUsecase *silverfish.SessionUsecase
	router         interf.IRouter
	route          string
}

// NewBlueprintUser export
func NewBlueprintUser(silverfish *silverfish.Silverfish, router interf.IRouter, sessionUsecase *silverfish.SessionUsecase) *BlueprintUser {
	bpu := new(BlueprintUser)
	bpu.silverfish = silverfish
	bpu.auth = silverfish.Auth
	bpu.route = "/user"
	bpu.router = router
	bpu.sessionUsecase = sessionUsecase
	return bpu
}

// RouteRegister export
func (bpu *BlueprintUser) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix(bpu.route).Subrouter()
	router.HandleFunc("", bpu.root)
	router.HandleFunc("/", bpu.root)
	router.HandleFunc("/status", bpu.status).Methods("GET")
	router.HandleFunc("/register", bpu.register).Methods("POST")
	router.HandleFunc("/login", bpu.login).Methods("POST")
	router.HandleFunc("/logout", bpu.logout).Methods("GET")
	router.HandleFunc("/isAdmin", bpu.isAdmin).Methods("GET")
	router.HandleFunc("/bookmark", bpu.bookmark).Methods("GET")

	router.HandleFunc("s/status", bpu.status).Methods("GET")
	router.HandleFunc("s/register", bpu.register).Methods("POST")
	router.HandleFunc("s/login", bpu.login).Methods("POST")
	router.HandleFunc("s/logout", bpu.logout).Methods("GET")
	router.HandleFunc("s/isAdmin", bpu.isAdmin).Methods("GET")
	router.HandleFunc("s/bookmark", bpu.bookmark).Methods("GET")

	authRouter := parentRouter.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/status", bpu.status).Methods("GET")
	authRouter.HandleFunc("/register", bpu.register).Methods("POST")
	authRouter.HandleFunc("/login", bpu.login).Methods("POST")
	authRouter.HandleFunc("/logout", bpu.logout).Methods("GET")
	authRouter.HandleFunc("/isAdmin", bpu.isAdmin).Methods("GET")
}

func (bpu *BlueprintUser) root(w http.ResponseWriter, r *http.Request) {

}

func (bpu *BlueprintUser) status(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Authorization")
	w.Header().Set("Content-Type", "application/json")

	result := bpu.sessionUsecase.IsTokenValid(&sessionToken)
	response := &entity.APIResponse{
		Success: true,
		Data:    result,
	}
	js, _ := json.Marshal(response)
	w.Write(js)
}

func (bpu *BlueprintUser) register(w http.ResponseWriter, r *http.Request) {
	account := r.FormValue("account")
	password := r.FormValue("password")
	recaptchaToken := r.FormValue("recaptchaToken")
	w.Header().Set("Content-Type", "application/json")

	response := new(entity.APIResponse)
	if recaptchaToken == "" {
		response = entity.NewAPIResponse(nil, errors.New("Missing recaptcha token"))
	} else if res, err := bpu.router.VerifyRecaptcha(&recaptchaToken); res == false {
		fmt.Println(err)
		response = entity.NewAPIResponse(nil, errors.New("Recaptcha verify failed"))
	} else {
		user, err := bpu.auth.Register(false, &account, &password)
		response = entity.NewAPIResponse(map[string]interface{}{
			"session": *bpu.sessionUsecase.InsertSession(user, false),
			"user":    user,
		}, err)
	}
	js, _ := json.Marshal(response)
	w.Write(js)
}

func (bpu *BlueprintUser) login(w http.ResponseWriter, r *http.Request) {
	keepLogin := r.FormValue("keepLogin")
	account := r.FormValue("account")
	password := r.FormValue("password")
	recaptchaToken := r.FormValue("recaptchaToken")
	w.Header().Set("Content-Type", "application/json")

	response := new(entity.APIResponse)
	if recaptchaToken == "" {
		response = entity.NewAPIResponse(nil, errors.New("Missing recaptcha token"))
	} else if res, err := bpu.router.VerifyRecaptcha(&recaptchaToken); res == false {
		fmt.Println(err)
		response = entity.NewAPIResponse(nil, errors.New("Recaptcha verify failed"))
	} else {
		user, err := bpu.auth.Login(&account, &password)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else {
			response = entity.NewAPIResponse(
				map[string]interface{}{
					"session": *bpu.sessionUsecase.InsertSession(user, keepLogin == "true"),
					"user":    user,
				}, nil)
		}
	}
	js, _ := json.Marshal(response)
	w.Write(js)
}

func (bpu *BlueprintUser) logout(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Authorization")
	w.Header().Set("Content-Type", "application/json")

	bpu.sessionUsecase.KillSession(&sessionToken)
	response := entity.NewAPIResponse(nil, nil)
	js, _ := json.Marshal(response)
	w.Write(js)
}

func (bpu *BlueprintUser) isAdmin(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Authorization")
	w.Header().Set("Content-Type", "application/json")

	session, err := bpu.sessionUsecase.GetSession(&sessionToken)
	response := new(entity.APIResponse)
	if err != nil {
		response = entity.NewAPIResponse(nil, err)
	} else {
		isAdmin, _ := bpu.auth.IsAdmin(session.GetAccount())
		response = entity.NewAPIResponse(map[string]interface{}{
			"isAdmin": isAdmin,
		}, nil)
	}
	js, _ := json.Marshal(response)
	w.Write(js)
}

func (bpu *BlueprintUser) bookmark(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Authorization")
	w.Header().Set("Content-Type", "application/json")

	session, err := bpu.sessionUsecase.GetSession(&sessionToken)
	response := new(entity.APIResponse)
	if err != nil {
		response = entity.NewAPIResponse(nil, err)
	} else {
		result, err := bpu.auth.GetUserBookmark(session.GetAccount())
		response = entity.NewAPIResponse(result, err)
	}
	js, _ := json.Marshal(response)
	w.Write(js)
}
