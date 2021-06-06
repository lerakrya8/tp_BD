package repository

import (
	"BD-v2/internal/app/users/models"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UsersRepository struct {
	DBPool *pgxpool.Pool
}

func NewUsersRepository(db *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{
		DBPool: db,
	}
}

func (rep *UsersRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `insert into users (nickname, fullname, about, email)
	VALUES ($1, $2, $3, $4)`

	_, err := rep.DBPool.Exec(ctx, query, user.Nickname, user.FullName,
		user.About, user.Email)
	return err
}

func (rep *UsersRepository) CheckIfUserExist(ctx context.Context,
	user *models.User) ([]*models.User, error) {
	query := `select nickname, fullname, about, email from users
	where nickname = $1 or email = $2`

	rows, err := rep.DBPool.Query(ctx, query, user.Nickname, user.Email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	existedUsers := make([]*models.User, 0)

	for rows.Next() {
		user = &models.User{}
		_ = rows.Scan(&user.Nickname, &user.FullName, &user.About, &user.Email)

		existedUsers = append(existedUsers, user)
	}
	return existedUsers, nil
}

func (rep *UsersRepository) FindUserNickname(ctx context.Context, nickname string) (*models.User, error) {
	query := `select nickname, fullname, about, email from users
		where nickname = $1`
	user := &models.User{}

	err := rep.DBPool.QueryRow(ctx, query, nickname).Scan(&user.Nickname, &user.FullName,
		&user.About, &user.Email)
	return user, err
}

func (rep *UsersRepository) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	query := `update users set fullname = COALESCE(NULLIF($1, ''), fullname),
		email = COALESCE(NULLIF($2, ''), email),
		about = COALESCE(NULLIF($3, ''), about)
		where nickname = $4 returning nickname, fullname, about, email`

	err := rep.DBPool.QueryRow(ctx, query, user.FullName, user.Email,
		user.About, user.Nickname).Scan(&user.Nickname, &user.FullName, &user.About, &user.Email)
	return user, err
}

func (rep *UsersRepository) GetUsers(limit int, forum, since string, desc bool) ([]*models.User, error) {
	query := `select u.nickname, u.fullname, u.about, u.email from users u
left join users_forum uf on u.nickname = uf.nickname
where uf.slug = $1
		`
	if since != "" && desc {
		query += fmt.Sprintf(" and u.nickname < '%s'", since)
	} else if since != "" {
		query += fmt.Sprintf(" and u.nickname > '%s'", since)
	}

	if desc {
		query += " order by u.nickname desc"
	} else {
		query += " order by u.nickname"
	}

	query += fmt.Sprintf(" limit NULLIF(%d, 0)", limit)

	users := make([]*models.User, 0)
	rows, err := rep.DBPool.Query(context.Background(), query, forum)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := &models.User{}
		err = rows.Scan(&user.Nickname, &user.FullName, &user.About, &user.Email)
		users = append(users, user)
	}
	return users, err
}
