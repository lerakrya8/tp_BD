package repository

import (
	"BD-v2/internal/app/forums/models"
	models2 "BD-v2/internal/app/threads/models"
	models3 "BD-v2/internal/app/users/models"
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/jmoiron/sqlx"
	"time"
)

type ForumRepository struct {
	db *sqlx.DB
}

func NewForumRepository(db *sqlx.DB) *ForumRepository {
	return &ForumRepository{
		db: db,
	}
}

func (rep *ForumRepository) CreateForum(forum *models.Forum) error {
	query := `insert into forums (title, users, slug) values ($1, $2, $3)`

	_, err := rep.db.Exec(query, forum.Title, forum.User, forum.Slug)
	return err
}

func (rep *ForumRepository) GetForumSlug(slug string) (*models.Forum, error) {
	query := `select title, users, slug, posts, threads from forums
		where slug = $1`
	forum := &models.Forum{}
	err := rep.db.Get(forum, query, slug)

	return forum, err
}

func (rep *ForumRepository) GetTreads(limit int, forum, since string, desc bool) ([]*models2.Thread, error) {
	query := `select id, title, author, forum, message, votes, slug, created from threads
		where forum = $1
		`
	if since != "" && desc {
		query += fmt.Sprintf(" and created <= '%s'", since)
	} else if since != "" {
		query += fmt.Sprintf(" and created >= '%s'", since)
	}

	if desc {
		query += " order by created desc"
	} else {
		query += " order by created"
	}

	query += fmt.Sprintf(" limit NULLIF(%d, 0)", limit)

	threads := make([]*models2.Thread, 0)
	rows, err := rep.db.Query(query, forum)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		t := &time.Time{}
		thread := &models2.Thread{}
		err = rows.Scan(&thread.ID, &thread.Title,
			&thread.Author, &thread.Forum, &thread.Message, &thread.Votes,
			&thread.Slug, t)
		thread.Created = strfmt.DateTime(t.UTC()).String()
		threads = append(threads, thread)
	}
	return threads, err
}

func (rep *ForumRepository) GetStatus() *models.Status {
	query1 := `select count(*) from forums`
	query2 := `select count(*) from posts`
	query3 := `select count(*) from threads`
	query4 := `select count(*) from users`

	status := &models.Status{}
	_ = rep.db.QueryRow(query1).Scan(&status.ForumsAmount)
	_ = rep.db.QueryRow(query2).Scan(&status.PostsAmount)
	_ = rep.db.QueryRow(query3).Scan(&status.ThreadsAmount)
	_ = rep.db.QueryRow(query4).Scan(&status.UsersAmount)
	return status
}

func (rep *ForumRepository) ClearDB() error {
	query1 := `truncate table thread_votes cascade `
	query2 := `truncate table posts cascade`
	query3 := `truncate table thread_votes cascade`
	query4 := `truncate table threads cascade `
	query5 := `truncate table users cascade`

	_, _ = rep.db.Exec(query1)
	_, _ = rep.db.Exec(query2)
	_, _ = rep.db.Exec(query3)
	_, _ = rep.db.Exec(query4)
	_, _ = rep.db.Exec(query5)
	return nil
}

func (rep *ForumRepository) GetForumUsers(slug, since string, limit int, desc bool) ([]*models3.User, error) {
	query := `select nickname, fullname, about, email from users
		left join forums f on users.nickname = f.users
		where f.slug = $1`

	if desc {
		query += fmt.Sprintf(` and users.nickname < '%s order by users.nickname desc'`, since)
	} else if since != "" {
		query += fmt.Sprintf(` and users.nickname > '%s' order by users.nickname`, since)
	}
	query += fmt.Sprintf(` limit %d`, limit)

	fmt.Println(query)

	rows, err := rep.db.Query(query, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	existedUsers := make([]*models3.User, 0)

	for rows.Next() {
		user := &models3.User{}
		_ = rows.Scan(&user.Nickname, &user.FullName, &user.About, &user.Email)

		existedUsers = append(existedUsers, user)
	}
	return existedUsers, nil
}
