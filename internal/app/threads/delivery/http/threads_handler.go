package http

import (
	"BD-v2/internal/app/forums"
	models2 "BD-v2/internal/app/posts/models"
	"BD-v2/internal/app/threads"
	"BD-v2/internal/app/threads/models"
	"BD-v2/internal/app/users"
	allModels "BD-v2/internal/models"
	tools "BD-v2/pkg"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgconn"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type ThreadsHandler struct {
	threadRep threads.Repository
	forumRep  forums.Repository
	userRep   users.Repository
}

func NewThreadsHandler(r *mux.Router, threadRep threads.Repository,
	forumRep forums.Repository, userRep users.Repository) *ThreadsHandler {
	theadHandler := &ThreadsHandler{
		threadRep: threadRep,
		forumRep:  forumRep,
		userRep:   userRep,
	}

	r.HandleFunc("/api/thread/{slug}/create", theadHandler.CreatePosts).Methods(http.MethodPost)
	r.HandleFunc("/api/thread/{slug}/vote", theadHandler.AddVoice).Methods(http.MethodPost)
	r.HandleFunc("/api/forum/{slug}/create", theadHandler.CrateThread).Methods(http.MethodPost)
	r.HandleFunc("/api/forum/{slug}/threads", theadHandler.GetTreads).Methods(http.MethodGet)
	r.HandleFunc("/api/thread/{slug_or_id}/details", theadHandler.GetTread).Methods(http.MethodGet)
	r.HandleFunc("/api/thread/{slug_or_id}/details", theadHandler.UpdateThread).Methods(http.MethodPost)
	r.HandleFunc("/api/thread/{slug_or_id}/posts", theadHandler.GetPosts).Methods(http.MethodGet)
	r.HandleFunc("/api/post/{id}/details", theadHandler.GetOnePost).Methods(http.MethodGet)
	r.HandleFunc("/api/post/{id}/details", theadHandler.PostUpdate).Methods(http.MethodPost)

	return theadHandler
}

func (handler *ThreadsHandler) CrateThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	forum, ok := vars["slug"]
	if !ok {
		fmt.Println("не шмогли достать slug")
		w.WriteHeader(500)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Ошибка чтения body")
		return
	}
	thread := &models.Thread{}
	thread.UnmarshalJSON(body)
	forumModel, err := handler.forumRep.GetForumSlug(forum)
	if err != nil {
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут юзера найти %s", thread.Author),
		}.MarshalJSON()
		w.Write(resp)
		return
	}
	thread.Forum = forumModel.Slug
	threadSlug := thread.Slug
	thread, err = handler.threadRep.CreateThread(context.Background(), thread)
	if err != nil {
		if err, ok := err.(*pgconn.PgError); ok {
			if err.Code == "23505" {
				thr, _ := handler.threadRep.FindThreadSlug(context.Background(), threadSlug)
				respBody, _ := thr.MarshalJSON()
				w.WriteHeader(409)
				w.Write(respBody)
			} else {
				w.WriteHeader(404)
				resp, _ := allModels.FailedResponse{
					Message: fmt.Sprintf("Не могут юзера найти %s", thread.Author),
				}.MarshalJSON()
				w.Write(resp)
			}
		}
		return
	}
	respBody, _ := thread.MarshalJSON()
	w.WriteHeader(http.StatusCreated)
	w.Write(respBody)
}

func (handler *ThreadsHandler) GetTreads(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 1
	}
	desc := tools.ConvertToBool(r.URL.Query().Get("desc"))
	since := r.URL.Query().Get("since")
	slug, _ := (mux.Vars(r))["slug"]

	forum, err := handler.forumRep.GetForumSlug(slug)
	if err != nil {
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут найти форум %s", forum.Slug),
		}.MarshalJSON()
		w.Write(resp)
		return
	}

	threads, err := handler.forumRep.GetTreads(limit, slug, since, desc)
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

func (handler *ThreadsHandler) CreatePosts(w http.ResponseWriter, r *http.Request) {
	counter++
	if counter == 3 {
		fmt.Println(3)
	}
	posts := make([]*models2.Post, 0)
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &posts)
	vars := mux.Vars(r)
	slug, _ := vars["slug"]
	threadID, err := strconv.Atoi(slug)
	thread := &models.Thread{}
	ctx := context.Background()

	if err != nil {
		thread, err = handler.threadRep.FindThreadSlug(ctx, slug)
	} else {
		thread, err = handler.threadRep.FindThreadID(ctx, threadID)
	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут юзера найти %s", "fds"),
		}.MarshalJSON()
		w.Write(resp)
		return
	}

	if len(posts) == 0 {
		respBody, _ := json.Marshal(posts)
		w.WriteHeader(http.StatusCreated)
		w.Write(respBody)
		return
	}

	_, err = handler.userRep.FindUserNickname(context.Background(), posts[0].Author)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут юзера найти %s", "fds"),
		}.MarshalJSON()
		w.Write(resp)
		return
	}

	posts, err = handler.threadRep.CreatePost(posts, thread)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут юзера найти %s", "fds"),
		}.MarshalJSON()
		w.Write(resp)
		return
	}
	if len(posts) == 0 {
		fmt.Println(err)
		w.WriteHeader(409)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут юзера найти %s", "fds"),
		}.MarshalJSON()
		w.Write(resp)
		return
	}
	respBody, _ := json.Marshal(posts)
	w.WriteHeader(http.StatusCreated)
	w.Write(respBody)
}

var (
	counter1 = 0
)

func (handler *ThreadsHandler) AddVoice(w http.ResponseWriter, r *http.Request) {
	counter1++
	if counter1 == 408 {
		fmt.Println(408)
	}
	voice := &models.Vote{}
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &voice)

	vars := mux.Vars(r)
	slug := vars["slug"]
	threadID, err := strconv.Atoi(slug)
	thread := &models.Thread{}
	ctx := context.Background()
	if err != nil {
		thread, err = handler.threadRep.FindThreadSlug(ctx, slug)
	} else {
		thread, err = handler.threadRep.FindThreadID(ctx, threadID)
	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут юзера найти %s", "fds"),
		}.MarshalJSON()
		w.Write(resp)
		return
	}
	voice.ThreadID = thread.ID
	err = handler.threadRep.AddVoice(ctx, voice)
	//updated := false
	if err != nil {
		if sqlerr, ok := err.(*pgconn.PgError); ok {
			if sqlerr.Code == "23503" {
				fmt.Println(err)
				w.WriteHeader(404)
				resp, _ := allModels.FailedResponse{
					Message: fmt.Sprintf("Не могут юзера найти %s", "fds"),
				}.MarshalJSON()
				w.Write(resp)
				return
			}
		}
		_, err := handler.threadRep.UpdateVoice(ctx, voice)
		if err != nil {
			fmt.Print(err)
			respBody, _ := json.Marshal(thread)
			w.WriteHeader(http.StatusOK)
			w.Write(respBody)
			return
		}
		//if isUpdated != 0 {
		//	updated = true
		//}
	}
	//if updated {
	//	thread.Votes += voice.Voice * 2
	//} else {
	//	thread.Votes += voice.Voice
	//}
	newThread, _ := handler.threadRep.FindThreadID(ctx, thread.ID)
	thread.Votes = newThread.Votes
	respBody, _ := json.Marshal(thread)
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func (handler *ThreadsHandler) GetTread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug_or_id"]
	threadID, err := strconv.Atoi(slug)
	thread := &models.Thread{}
	ctx := context.Background()
	if err != nil {
		thread, err = handler.threadRep.FindThreadSlug(ctx, slug)
	} else {
		thread, err = handler.threadRep.FindThreadID(ctx, threadID)
	}

	if err != nil {
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут найти форум %s", slug),
		}.MarshalJSON()
		w.Write(resp)
		return
	}
	respBody, _ := json.Marshal(thread)
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func (handler *ThreadsHandler) UpdateThread(w http.ResponseWriter, r *http.Request) {
	thread := &models.Thread{}
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &thread)

	vars := mux.Vars(r)
	slug := vars["slug_or_id"]
	threadID, err := strconv.Atoi(slug)
	ctx := context.Background()
	prevThread := &models.Thread{}
	if err != nil {
		prevThread, err = handler.threadRep.FindThreadSlug(ctx, slug)
		if err != nil {
			w.WriteHeader(404)
			resp, _ := allModels.FailedResponse{
				Message: fmt.Sprintf("Не могут найти форум %s", slug),
			}.MarshalJSON()
			w.Write(resp)
			return
		}
		if thread.Title != "" {
			prevThread.Title = thread.Title
		}
		if thread.Message != "" {
			prevThread.Message = thread.Message
		}
		err = handler.threadRep.UpdateThreadSlug(ctx, prevThread)
		if err != nil {
			w.WriteHeader(404)
			resp, _ := allModels.FailedResponse{
				Message: fmt.Sprintf("Не могут найти форум %s", slug),
			}.MarshalJSON()
			w.Write(resp)
			return
		}
	} else {
		prevThread, err = handler.threadRep.FindThreadID(ctx, threadID)
		if err != nil {
			w.WriteHeader(404)
			resp, _ := allModels.FailedResponse{
				Message: fmt.Sprintf("Не могут найти форум %s", slug),
			}.MarshalJSON()
			w.Write(resp)
			return
		}
		if thread.Title != "" {
			prevThread.Title = thread.Title
		}
		if thread.Message != "" {
			prevThread.Message = thread.Message
		}
		err = handler.threadRep.UpdateThreadID(ctx, prevThread)
		if err != nil {
			w.WriteHeader(404)
			resp, _ := allModels.FailedResponse{
				Message: fmt.Sprintf("Не могут найти форум %s", slug),
			}.MarshalJSON()
			w.Write(resp)
			return
		}
	}

	respBody, _ := json.Marshal(prevThread)
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

var (
	counter = 0
)

func (handler *ThreadsHandler) GetPosts(w http.ResponseWriter, r *http.Request) {

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 100
	}
	descSTR := r.URL.Query().Get("desc")
	if descSTR == "" {
		fmt.Println("")
	}
	desc := tools.ConvertToBool(descSTR)
	since, err := strconv.Atoi(r.URL.Query().Get("since"))
	if err != nil {
		since = 0
	}
	sort := r.URL.Query().Get("sort")

	vars := mux.Vars(r)
	slug := vars["slug_or_id"]
	threadID, err := strconv.Atoi(slug)
	thread := &models.Thread{}
	ctx := context.Background()
	if err != nil {
		thread, err = handler.threadRep.FindThreadSlug(ctx, slug)
	} else {
		thread, err = handler.threadRep.FindThreadID(ctx, threadID)
	}
	if err != nil {
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут найти форум %s", slug),
		}.MarshalJSON()
		w.Write(resp)
		return
	}
	if thread.ID != 0 {
		threadID = thread.ID
	}

	var posts []*models2.Post
	if sort == "flat" || sort == "" {
		posts, err = handler.threadRep.GetPosts(threadID, limit, since, desc)
		if err != nil {
			w.WriteHeader(404)
			resp, _ := allModels.FailedResponse{
				Message: fmt.Sprintf("Не могут найти форум %s", slug),
			}.MarshalJSON()
			w.Write(resp)
			return
		}
	} else if sort == "tree" {
		posts, err = handler.threadRep.GetPostsTree(threadID, limit, since, desc)
		if err != nil {
			w.WriteHeader(404)
			resp, _ := allModels.FailedResponse{
				Message: fmt.Sprintf("Не могут найти форум %s", slug),
			}.MarshalJSON()
			w.Write(resp)
			return
		}
	} else if sort == "parent_tree" {
		posts, err = handler.threadRep.GetPostsParentTree(threadID, limit, since, desc)
		if err != nil {
			w.WriteHeader(404)
			resp, _ := allModels.FailedResponse{
				Message: fmt.Sprintf("Не могут найти форум %s", slug),
			}.MarshalJSON()
			w.Write(resp)
			return
		}
	}

	if posts == nil {
		posts = make([]*models2.Post, 0)
	}

	respBody, _ := json.Marshal(posts)
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func (handler *ThreadsHandler) GetOnePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["id"]
	postID, _ := strconv.Atoi(slug)
	args := r.URL.Query().Get("related")
	argsMas := strings.Split(args, ",")
	for len(argsMas) < 3 {
		argsMas = append(argsMas, " ")
	}

	postRelated, err := handler.threadRep.GetOnePost(context.Background(), postID, argsMas)
	if err != nil {
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут найти форум %s", slug),
		}.MarshalJSON()
		w.Write(resp)
		return
	}

	respBody, _ := json.Marshal(postRelated)
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func (handler *ThreadsHandler) PostUpdate(w http.ResponseWriter, r *http.Request) {
	post := &models2.Post{}
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &post)
	vars := mux.Vars(r)
	slug := vars["id"]
	postID, _ := strconv.Atoi(slug)

	prevPost, err := handler.threadRep.GetPostByID(context.Background(), postID)
	if err != nil {
		w.WriteHeader(404)
		resp, _ := allModels.FailedResponse{
			Message: fmt.Sprintf("Не могут найти форум %s", slug),
		}.MarshalJSON()
		w.Write(resp)
		return
	}

	post.ID = postID
	if prevPost.Message != post.Message && post.Message != "" {
		err = handler.threadRep.UpdatePost(context.Background(), post)
		if err != nil {
			w.WriteHeader(404)
			resp, _ := allModels.FailedResponse{
				Message: fmt.Sprintf("Не могут найти форум %s", slug),
			}.MarshalJSON()
			w.Write(resp)
			return
		}
		prevPost.Message = post.Message
		prevPost.ISEdited = true
	}

	respBody, _ := json.Marshal(prevPost)
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}
