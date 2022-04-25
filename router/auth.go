package router

import (
	"encoding/json"
	"errors"
	"net/http"
	interf "silverfish/router/interface"
	"silverfish/silverfish"
	"silverfish/silverfish/entity"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// BlueprintAuth export
type BlueprintAuth struct {
	auth   *silverfish.Auth
	router interf.IRouter
	route  string
}

// NewBlueprintAuth export
func NewBlueprintAuth(
	auth *silverfish.Auth,
	router interf.IRouter,
) *BlueprintAuth {
	bpa := new(BlueprintAuth)
	bpa.auth = auth
	bpa.route = "/auth"
	bpa.router = router
	return bpa
}

// RouterRegister export
func (bpa *BlueprintAuth) RouterRegiter(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix(bpa.route).Subrouter()
	router.HandleFunc("/status", bpa.status).Methods("GET")
	router.HandleFunc("/register", bpa.register).Methods("POST")
	router.HandleFunc("/login", bpa.login).Methods("POST")
	router.HandleFunc("/logout", bpa.logout).Methods("GET")
	router.HandleFunc("/isAdmin", bpa.isAdmin).Methods("GET")
}

func (bpa *BlueprintAuth) status(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Authorization")
	w.Header().Set("Content-Type", "application/json")

	result := bpa.auth.IsTokenValid(&sessionToken)
	response := &entity.APIResponse{
		Success: true,
		Data:    result,
	}
	js, _ := json.Marshal(response)
	w.Write(js)
}

func (bpa *BlueprintAuth) register(w http.ResponseWriter, r *http.Request) {
	account := r.FormValue("account")
	password := r.FormValue("password")
	recaptchaToken := r.FormValue("recaptchaToken")
	w.Header().Set("Content-Type", "application/json")

	response := new(entity.APIResponse)
	if recaptchaToken == "" {
		response = entity.NewAPIResponse(nil, errors.New("Missing recaptcha token"))
	} else if res, err := bpa.router.VerifyRecaptcha(&recaptchaToken); res == false {
		logrus.Info(err)
		response = entity.NewAPIResponse(nil, errors.New("Recaptcha verify failed"))
	} else {
		user, err := bpa.auth.Register(false, &account, &password)
		response = entity.NewAPIResponse(map[string]interface{}{
			"session": *bpa.auth.InsertSession(user, false),
			"user":    user,
		}, err)
	}
	js, _ := json.Marshal(response)
	w.Write(js)
}

func (bpa *BlueprintAuth) login(w http.ResponseWriter, r *http.Request) {
	keepLogin := r.FormValue("keepLogin")
	account := r.FormValue("account")
	password := r.FormValue("password")
	recaptchaToken := r.FormValue("recaptchaToken")
	w.Header().Set("Content-Type", "application/json")

	response := new(entity.APIResponse)
	if recaptchaToken == "" {
		response = entity.NewAPIResponse(nil, errors.New("Missing recaptcha token"))
	} else if res, err := bpa.router.VerifyRecaptcha(&recaptchaToken); res == false {
		logrus.Info(err)
		response = entity.NewAPIResponse(nil, errors.New("Recaptcha verify failed"))
	} else {
		user, err := bpa.auth.Login(&account, &password)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else {
			session := *bpa.auth.InsertSession(user, keepLogin == "true")
			sessionRtn := map[string]interface{}{
				"token":          session.GetToken(),
				"expireDatetime": session.GetExpireTS(),
			}
			response = entity.NewAPIResponse(
				map[string]interface{}{
					"session": sessionRtn,
					"user":    user,
				}, nil)
		}
	}
	js, _ := json.Marshal(response)
	w.Write(js)
}

func (bpa *BlueprintAuth) logout(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Authorization")
	w.Header().Set("Content-Type", "application/json")

	bpa.auth.KillSession(&sessionToken)
	response := entity.NewAPIResponse(nil, nil)
	js, _ := json.Marshal(response)
	w.Write(js)
}

func (bpa *BlueprintAuth) isAdmin(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Authorization")
	w.Header().Set("Content-Type", "application/json")

	session, err := bpa.auth.GetSession(&sessionToken)
	response := new(entity.APIResponse)
	if err != nil {
		response = entity.NewAPIResponse(nil, err)
	} else {
		isAdmin, _ := bpa.auth.IsAdmin(session.GetAccount())
		response = entity.NewAPIResponse(map[string]interface{}{
			"isAdmin": isAdmin,
		}, nil)
	}
	js, _ := json.Marshal(response)
	w.Write(js)
}
