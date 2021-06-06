package repository

import (
	models4 "BD-v2/internal/app/forums/models"
	models2 "BD-v2/internal/app/posts/models"
	models3 "BD-v2/internal/app/posts/post_related"
	"BD-v2/internal/app/threads/models"
	models1 "BD-v2/internal/app/users/models"
	"context"
	"database/sql"
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"strings"
	"time"
)

type ThreadsRepository struct {
	DBPool *pgxpool.Pool
}

func NewThreadsRepository(db *pgxpool.Pool) *ThreadsRepository {
	return &ThreadsRepository{
		DBPool: db,
	}
}

func (rep *ThreadsRepository) CreateThread(ctx context.Context, thread *models.Thread) (*models.Thread, error) {
	var query string
	var err error
	if thread.Slug != "" {
		if thread.Created != "" {
			query = `insert into threads (title, author, forum, message, slug, created)
			values ($1, $2, $3, $4, $5, $6) returning id, title, author, forum,
			message, votes, slug`
			err = rep.DBPool.QueryRow(ctx, query, thread.Title, thread.Author,
				thread.Forum, thread.Message, thread.Slug, thread.Created).Scan(&thread.ID, &thread.Title,
				&thread.Author, &thread.Forum, &thread.Message, &thread.Votes,
				&thread.Slug,
			)
		} else {
			query = `insert into threads (title, author, forum, message, slug)
			values ($1, $2, $3, $4, $5) returning id, title, author, forum,
			message, votes, slug`
			err = rep.DBPool.QueryRow(ctx, query, thread.Title, thread.Author,
				thread.Forum, thread.Message, thread.Slug).Scan(&thread.ID, &thread.Title,
				&thread.Author, &thread.Forum, &thread.Message, &thread.Votes,
				&thread.Slug)
		}
	} else {
		if thread.Created != "" {
			query = `insert into threads (title, author, forum, message, created)
			values ($1, $2, $3, $4, $5) returning id, title, author, forum,
			message, votes`
			err = rep.DBPool.QueryRow(ctx, query, thread.Title, thread.Author,
				thread.Forum, thread.Message, &thread.Created).Scan(&thread.ID, &thread.Title,
				&thread.Author, &thread.Forum, &thread.Message, &thread.Votes,
			)
		} else {
			query = `insert into threads (title, author, forum, message)
			values ($1, $2, $3, $4) returning id, title, author, forum,
			message, votes`
			err = rep.DBPool.QueryRow(ctx, query, thread.Title, thread.Author,
				thread.Forum, thread.Message).Scan(&thread.ID, &thread.Title,
				&thread.Author, &thread.Forum, &thread.Message, &thread.Votes,
			)
		}
	}
	return thread, err
}

func (rep *ThreadsRepository) FindThreadSlug(ctx context.Context, slug string) (*models.Thread, error) {
	query := `select id, title, author, forum,
	message, votes, slug, created from threads
	where slug = $1`
	thread := &models.Thread{}
	date := time.Time{}
	nullSlug := sql.NullString{}

	err := rep.DBPool.QueryRow(ctx, query, slug).Scan(&thread.ID, &thread.Title, &thread.Author,
		&thread.Forum, &thread.Message, &thread.Votes, &nullSlug, &date,
	)

	if nullSlug.Valid {
		thread.Slug = nullSlug.String
	}

	thread.Created = strfmt.DateTime(date.UTC()).String()
	return thread, err
}

func (rep *ThreadsRepository) FindThreadID(ctx context.Context, threadID int) (*models.Thread, error) {
	query := `select id, title, author, forum,
	message, votes, slug, created from threads
	where id = $1`

	thread := &models.Thread{}
	date := time.Time{}
	nullSlug := sql.NullString{}

	err := rep.DBPool.QueryRow(ctx, query, threadID).Scan(&thread.ID, &thread.Title, &thread.Author,
		&thread.Forum, &thread.Message, &thread.Votes, &nullSlug, &date,
	)

	if nullSlug.Valid {
		thread.Slug = nullSlug.String
	}
	thread.Created = strfmt.DateTime(date.UTC()).String()
	return thread, err
}

func (rep *ThreadsRepository) CreatePost(posts []*models2.Post, thread *models.Thread) ([]*models2.Post, error) {
	query := `INSERT INTO posts(author, created, message, parent, thread, forum) VALUES `
	empty := make([]*models2.Post, 0)
	if len(posts) == 0 {
		return empty, nil
	}

	created := time.Now()
	names := make([]string, 0)
	params := make([]interface{}, 0)
	i := 1
	for _, item := range posts {
		names = append(names, fmt.Sprintf(
			" ($%d, $%d, $%d, $%d, $%d, $%d) ", i, i+1, i+2, i+3, i+4, i+5))
		i += 6
		params = append(params, item.Author, created, item.Message, item.Parent, thread.ID, thread.Forum)
	}
	query += strings.Join(names, ",")
	query += " RETURNING author, created, forum, id, is_edited, message, parent, thread"

	row, err := rep.DBPool.Query(context.Background(), query, params...)
	if err != nil {
		return empty, err
	}

	defer row.Close()

	for row.Next() {
		post := &models2.Post{}
		err = row.Scan(&post.Author, &created, &post.Forum, &post.ID, &post.ISEdited,
			&post.Message, &post.Parent, &post.Thread)
		if err != nil {
			return empty, err
		}

		post.Created = strfmt.DateTime(created.UTC()).String()
		empty = append(empty, post)
	}
	return empty, err
}

func (rep *ThreadsRepository) GetPosts(threadID, limit, since int, desc bool) ([]*models2.Post, error) {
	query := `SELECT id, author, message, is_edited, forum, thread, created, parent FROM posts WHERE 
	thread = $1 `

	formSorting(&query, desc, since)
	query += `limit nullif($2, 0)`

	posts := make([]*models2.Post, 0)

	row, err := rep.DBPool.Query(context.Background(), query, threadID, limit)

	if err != nil {
		return posts, err
	}
	defer row.Close()
	for row.Next() {
		post := &models2.Post{}
		created := &time.Time{}

		err = row.Scan(&post.ID, &post.Author, &post.Message, &post.ISEdited, &post.Forum,
			&post.Thread, created, &post.Parent)

		if err != nil {
			return posts, err
		}
		post.Created = strfmt.DateTime(created.UTC()).String()
		posts = append(posts, post)
	}
	return posts, err
}

func formSorting(query *string, desc bool, since int) {
	if desc {
		if since > 0 {
			*query += fmt.Sprintf("AND id < %d ", since)
		}
		*query += `ORDER BY id DESC `
	} else {
		if since > 0 {
			*query += fmt.Sprintf("AND id > %d ", since)
		}
		*query += `ORDER BY id `
	}
}

func (rep *ThreadsRepository) GetPostsTree(threadID, limit, since int, desc bool) ([]*models2.Post, error) {
	var query string
	prevQuey := ""
	if since != 0 {
		if desc {
			prevQuey = `AND path < `
		} else {
			prevQuey = `AND path > `
		}
		prevQuey += fmt.Sprintf(`(SELECT path FROM posts WHERE id = %d)`, since)
	}

	query = getMainQuery(prevQuey, desc)

	posts := make([]*models2.Post, 0)
	row, err := rep.DBPool.Query(context.Background(), query, threadID, limit)

	if err != nil {
		return posts, err
	}
	defer row.Close()

	for row.Next() {
		post := &models2.Post{}
		created := &time.Time{}

		err = row.Scan(&post.ID, &post.Author, &post.Message, &post.ISEdited, &post.Forum,
			&post.Thread, created, &post.Parent)

		if err != nil {
			return posts, err
		}
		post.Created = strfmt.DateTime(created.UTC()).String()
		posts = append(posts, post)

	}
	return posts, err
}

func getMainQuery(prevQuey string, desc bool) string {
	query := "SELECT id, author, message, is_edited, forum, thread, created, parent FROM posts \nWHERE thread=$1 "
	if desc {
		query += prevQuey + "ORDER BY path DESC, id DESC LIMIT NULLIF($2, 0);"
	} else {
		query += prevQuey + "ORDER BY path, id LIMIT NULLIF($2, 0);"
	}
	return query
}

func (rep *ThreadsRepository) GetPostsParentTree(threadID, limit, since int, desc bool) ([]*models2.Post, error) {
	prevQuery := ""
	if since != 0 {
		if desc {
			prevQuery = `AND PATH[1] < `
		} else {
			prevQuery = `AND PATH[1] > `
		}
		prevQuery += fmt.Sprintf(`(SELECT path[1] FROM posts WHERE id = %d)`, since)
	}

	query := parentTreeMainQuery(desc, prevQuery, limit)
	posts := make([]*models2.Post, 0)
	row, err := rep.DBPool.Query(context.Background(), query, threadID)
	if err != nil {
		return posts, err
	}

	defer row.Close()

	for row.Next() {
		post := &models2.Post{}
		created := &time.Time{}

		err = row.Scan(&post.ID, &post.Author, &post.Message, &post.ISEdited, &post.Forum,
			&post.Thread, created, &post.Parent)
		if err != nil {
			return posts, err
		}

		post.Created = strfmt.DateTime(created.UTC()).String()
		posts = append(posts, post)

	}
	return posts, err
}

func parentTreeMainQuery(desc bool, prevQuery string, limit int) string {
	query := ""
	findParentQuery := "SELECT id FROM posts WHERE thread = $1 AND parent IS NULL " + prevQuery + "ORDER BY id"
	if desc {
		findParentQuery += " DESC"
		if limit > 0 {
			findParentQuery += fmt.Sprintf(` LIMIT %d`, limit)
		}
		query = "SELECT id, author, message, is_edited, forum, thread, created, parent FROM posts WHERE path[1] IN (" +
			findParentQuery +  ") ORDER BY path[1] DESC, path, id"
	} else {
		if limit > 0 {
			findParentQuery += fmt.Sprintf(` LIMIT %d`, limit)
		}
		query = "SELECT id, author, message, is_edited, forum, thread, created, parent FROM posts WHERE path[1] IN (" +
			findParentQuery +  ") ORDER BY path,id"
	}
	return query
}

func (rep *ThreadsRepository) AddVoice(ctx context.Context, voice *models.Vote) error {
	query := `insert into thread_votes (nickname, voice, thread_id) VALUES ($1, $2, $3)`

	_, err := rep.DBPool.Exec(ctx, query, voice.Nickname, voice.Voice, voice.ThreadID)
	return err
}

func (rep *ThreadsRepository) GetVoice(ctx context.Context, voice *models.Vote) (*models.Vote, error) {
	query := `select nickname, voice, thread_id from thread_votes
		where nickname = $1 and thread_id = $2`

	err := rep.DBPool.QueryRow(ctx, query, voice.Nickname, voice.ThreadID).Scan(&voice.Nickname,
		&voice.Voice, &voice.ThreadID)
	return voice, err
}

func (rep *ThreadsRepository) UpdateVoice(ctx context.Context, voice *models.Vote) (int, error) {
	query := `update thread_votes
		set voice = $1
		where nickname = $2 and thread_id = $3 and voice != $1`

	res, err := rep.DBPool.Exec(ctx, query, voice.Voice, voice.Nickname, voice.ThreadID)
	if err != nil {
		return 0, err
	}
	return int(res.RowsAffected()), err
}

func (rep *ThreadsRepository) UpdateThreadID(ctx context.Context, thread *models.Thread) error {
	query := `update threads set title = $1, message = $2
	where id = $3
`

	_, err := rep.DBPool.Exec(ctx, query, thread.Title, thread.Message, thread.ID)
	return err
}

func (rep *ThreadsRepository) UpdateThreadSlug(ctx context.Context, thread *models.Thread) error {
	query := `update threads set title = $1, message = $2
	where slug = $3
`
	_, err := rep.DBPool.Exec(ctx, query, thread.Title, thread.Message, thread.Slug)
	return err
}

func (rep *ThreadsRepository) GetOnePost(ctx context.Context, postID int, rel []string) (*models3.PostRelated, error) {
	query := `SELECT id, author, message, is_edited, forum, thread, created, parent FROM posts
	WHERE id = $1`
	postRelated := &models3.PostRelated{}
	post := &models2.Post{}
	created := &time.Time{}

	err := rep.DBPool.QueryRow(ctx, query, postID).Scan(&post.ID, &post.Author, &post.Message,
		&post.ISEdited, &post.Forum, &post.Thread, created, &post.Parent)
	post.Created = strfmt.DateTime(created.UTC()).String()
	if err != nil {
		return nil, err
	}
	postRelated.Post = post

	if Contains(rel, "user") {
		user := &models1.User{}
		query = `select nickname, fullname, about, email from users
		where nickname = $1`

		err = rep.DBPool.QueryRow(ctx, query, post.Author).Scan(&user.Nickname,
			&user.FullName, &user.About, &user.Email,
		)
		if err != nil {
			return nil, err
		}
		postRelated.Author = user
	}

	if Contains(rel, "forum") {
		forum := &models4.Forum{}
		query = `select title, users, slug, posts, threads from forums
		where slug = $1`

		err = rep.DBPool.QueryRow(ctx, query, post.Forum).Scan(&forum.Title, &forum.User, &forum.Slug,
			&forum.Posts, &forum.Threads,
		)
		if err != nil {
			return nil, err
		}
		postRelated.Forum = forum
	}

	if Contains(rel, "thread") {
		thread, err := rep.FindThreadID(ctx, post.Thread)
		if err != nil {
			return nil, err
		}
		postRelated.Thread = thread
	}

	return postRelated, nil
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func (rep *ThreadsRepository) GetPostByID(ctx context.Context, postID int) (*models2.Post, error) {
	query := `SELECT id, author, message, is_edited, forum, thread, created, parent FROM posts
	WHERE id = $1`
	created := &time.Time{}
	post := &models2.Post{}

	err := rep.DBPool.QueryRow(ctx, query, postID).Scan(&post.ID, &post.Author, &post.Message,
		&post.ISEdited, &post.Forum, &post.Thread, created, &post.Parent)
	post.Created = strfmt.DateTime(created.UTC()).String()
	return post, err
}

func (rep *ThreadsRepository) UpdatePost(ctx context.Context, post *models2.Post) error {
	query := `update posts set message = $1, is_edited = true 
	where id = $2`

	_, err := rep.DBPool.Exec(ctx, query, post.Message, post.ID)
	return err
}
