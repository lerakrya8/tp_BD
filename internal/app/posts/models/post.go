package models

type Post struct {
	ID       int    `json:"id"`
	Parent   *int   `json:"parent,omitempty"`
	Author   string `json:"author"`
	Message  string `json:"message"`
	ISEdited bool   `json:"isEdited"`
	Forum    string `json:"forum"`
	Thread   int    `json:"thread"`
	Created  string `json:"created"`
}
