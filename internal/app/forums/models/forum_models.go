package models

type Forum struct {
	Title   string `json:"title" db:"title"`
	User    string `json:"user" db:"users"`
	Slug    string `json:"slug" db:"slug"`
	Posts   int    `json:"posts,omitempty" db:"posts"`
	Threads int    `json:"threads,omitempty" db:"threads"`
}

type Status struct {
	UsersAmount   int `json:"user"`
	ForumsAmount  int `json:"forum"`
	ThreadsAmount int `json:"thread"`
	PostsAmount   int `json:"post"`
}
