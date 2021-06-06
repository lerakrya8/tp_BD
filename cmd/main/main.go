package main

import (
	forumHandler "BD-v2/internal/app/forums/delivery/http"
	forumRep "BD-v2/internal/app/forums/repository"
	threadHandler "BD-v2/internal/app/threads/delivery/http"
	threadRep "BD-v2/internal/app/threads/repository"
	userHandler "BD-v2/internal/app/users/delivery/http"
	userRep "BD-v2/internal/app/users/repository"
	"BD-v2/internal/middlware"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jmoiron/sqlx"
	"net/http"
)

var (
	settings =  "host=localhost port=5432 user=lera_bd dbname=db password=password sslmode=disable"
	//settings = "host=localhost port=5432 dbname=lera_bd sslmode=disable"
)

func main() {
	pool, err := pgxpool.Connect(context.Background(), settings)
	db, err := sqlx.Connect("postgres", settings)
	if err != nil {
		fmt.Println("Не смогли подключиться к бд")
	}

	threadRep := threadRep.NewThreadsRepository(pool)
	forumRep := forumRep.NewForumRepository(db)
	userRep := userRep.NewUsersRepository(pool)
	router := mux.NewRouter()
	_ = userHandler.NewUsersHandler(router, userRep)
	_ = forumHandler.NewForumsHandler(router, forumRep, userRep)
	_ = threadHandler.NewThreadsHandler(router, threadRep, forumRep, userRep)

	router.Use(middlware.ContentType)
	http.ListenAndServe(":5000", router)
}
