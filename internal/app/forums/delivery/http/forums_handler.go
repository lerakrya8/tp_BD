package http

import (
	"BD-v2/internal/app/forums"
	"BD-v2/internal/app/forums/models"
	"BD-v2/internal/app/users"
	allModels "BD-v2/internal/models"
	tools "BD-v2/pkg"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"io/ioutil"
	"net/http"
	"strconv"
)

type ForumsHandler struct {
	forumsRep forums.Repository
	UserRep   users.Repository
}

func NewForumsHandler(r *mux.Router, rep forums.Repository, userRep users.Repository) *ForumsHandler {
	forumsHandler := &ForumsHandler{
		forumsRep: rep,
		UserRep:   userRep,
	}

	r.HandleFunc("/api/forum/create", forumsHandler.CreateForum).Methods(http.MethodPost)
	r.HandleFunc("/api/forum/{slug}/details", forumsHandler.GetForum).Methods(http.MethodGet)
	r.HandleFunc("/api/service/clear", forumsHandler.ClearDB).Methods(http.MethodPost)
	r.HandleFunc("/api/service/status", forumsHandler.GetStatus).Methods(http.MethodGet)
	r.HandleFunc("/api/forum/{slug}/users", forumsHandler.GetUsers).Methods(http.MethodGet)

	return forumsHandler
}

func (handler *ForumsHandler) CreateForum(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Ошибка чтения body")
		return
	}
	forum := &models.Forum{}
	err = forum.UnmarshalJSON(body)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	user, err := handler.UserRep.FindUserNickname(context.Background(), forum.User)
	if err != nil {
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут юзера найти %s", forum.User),
		}.MarshalJSON()
		w.Write(resp)
		return
	}
	forum.User = user.Nickname
	err = handler.forumsRep.CreateForum(forum)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				f, _ := handler.forumsRep.GetForumSlug(forum.Slug)
				respBody, _ := f.MarshalJSON()
				w.WriteHeader(409)
				w.Write(respBody)
			} else {
				w.WriteHeader(404)
				resp, _ := allModels.FailedResponse{
					Message: fmt.Sprintf("Не могут юзера найти %s", forum.User),
				}.MarshalJSON()
				w.Write(resp)
			}
		}
		return
	}
	respBody, _ := forum.MarshalJSON()
	w.WriteHeader(201)
	w.Write(respBody)
}

func (handler *ForumsHandler) GetForum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug, ok := vars["slug"]
	if !ok {
		fmt.Println("не шмогли достать slug")
		w.WriteHeader(500)
		return
	}
	forum, err := handler.forumsRep.GetForumSlug(slug)
	if err != nil {
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут найти форум %s", slug),
		}.MarshalJSON()
		w.Write(resp)
		return
	}
	respBody, _ := forum.MarshalJSON()
	w.WriteHeader(200)
	w.Write(respBody)
}

func (handler *ForumsHandler) ClearDB(w http.ResponseWriter, r *http.Request) {
	err := handler.forumsRep.ClearDB()
	if err != nil {
		w.WriteHeader(500)
	}
	w.WriteHeader(200)
}

func (handler *ForumsHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := handler.forumsRep.GetStatus()
	respBody, _ := status.MarshalJSON()
	w.WriteHeader(200)
	w.Write(respBody)
}

func (handler *ForumsHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 1
	}
	desc := tools.ConvertToBool(r.URL.Query().Get("desc"))
	since := r.URL.Query().Get("since")
	slug, _ := (mux.Vars(r))["slug"]

	forum, err := handler.forumsRep.GetForumSlug(slug)
	if err != nil {
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут найти форум %s", forum.Slug),
		}.MarshalJSON()
		w.Write(resp)
		return
	}

	threads, err := handler.UserRep.GetUsers(limit, slug, since, desc)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут юзера найти %s", "fds"),
		}.MarshalJSON()
		w.Write(resp)
		return
	}
	respBody, _ := json.Marshal(threads)
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}
