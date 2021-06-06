package http

import (
	"BD-v2/internal/app/users"
	"BD-v2/internal/app/users/models"
	allModels "BD-v2/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgconn"
	"io/ioutil"
	"net/http"
)

type UsersHandler struct {
	UsersRep users.Repository
}

func NewUsersHandler(r *mux.Router, rep users.Repository) *UsersHandler {
	usersHandler := &UsersHandler{
		UsersRep: rep,
	}

	r.HandleFunc("/api/user/{nickname}/create", usersHandler.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/api/user/{nickname}/profile", usersHandler.GetUser).Methods(http.MethodGet)
	r.HandleFunc("/api/user/{nickname}/profile", usersHandler.UpdateUser).Methods(http.MethodPost)

	return usersHandler
}


func (userHandler *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Ошибка чтения body")
		return
	}
	vars := mux.Vars(r)
	nickname, ok := vars["nickname"]
	if !ok {
		fmt.Println("не шмогли достать nickname")
		w.WriteHeader(500)
		return
	}
	user := &models.User{}
	user.Nickname = nickname
	err = user.UnmarshalJSON(body)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Ошибка распаковки json")
		return
	}

	err = userHandler.UsersRep.CreateUser(context.Background(), user)
	if err != nil {
		existedUsers, err := userHandler.UsersRep.CheckIfUserExist(context.Background(), user)
		if err != nil {
			w.WriteHeader(500)
			fmt.Println("какая-то проблемка")
			return
		}
		respBody, _ := json.Marshal(existedUsers)
		w.WriteHeader(http.StatusConflict)
		w.Write(respBody)
		return
	}
	respBody, _ := user.MarshalJSON()
	w.WriteHeader(http.StatusCreated)
	w.Write(respBody)
}

func (handler *UsersHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nickname, ok := vars["nickname"]
	if !ok {
		fmt.Println("не шмогли достать nickname")
		w.WriteHeader(500)
		return
	}

	user, err := handler.UsersRep.FindUserNickname(context.Background(), nickname)
	if err != nil {
		w.WriteHeader(404)
		fmt.Println("Нет такого юзера")
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Can't find user with nickname %s", nickname),
		}.MarshalJSON()
		w.Write(resp)
		return
	}
	w.WriteHeader(200)
	body, _ := user.MarshalJSON()
	w.Write(body)
}

func (userHandler *UsersHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Ошибка чтения body")
		return
	}
	vars := mux.Vars(r)
	nickname, ok := vars["nickname"]
	if !ok {
		fmt.Println("не шмогли достать nickname")
		w.WriteHeader(500)
		return
	}
	user := &models.User{}
	user.Nickname = nickname
	err = user.UnmarshalJSON(body)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Ошибка распаковки json")
		return
	}

	user, err = userHandler.UsersRep.UpdateUser(context.Background(), user)
	if err != nil {
		if err, ok := err.(*pgconn.PgError); ok {
			if err.Code == "23505" {
				w.WriteHeader(409)
				resp, _ := allModels.FailedResponse{
					Message: fmt.Sprintf("Уже существует %s", nickname),
				}.MarshalJSON()
				w.Write(resp)
				return
			}
			w.WriteHeader(404)
			resp, _ := allModels.FailedResponse{
				Message: fmt.Sprintf("Не могут юзера найти %s", nickname),
			}.MarshalJSON()
			w.Write(resp)
		} else {
			w.WriteHeader(404)
			resp, _ := allModels.FailedResponse{
				Message: fmt.Sprintf("Не могут юзера найти %s", nickname),
			}.MarshalJSON()
			w.Write(resp)
		}
		return
	}
	respBody, _ := user.MarshalJSON()
	w.WriteHeader(200)
	w.Write(respBody)
}
