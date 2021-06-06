package post_related

import (
	models4 "BD-v2/internal/app/forums/models"
	models2 "BD-v2/internal/app/posts/models"
	"BD-v2/internal/app/threads/models"
	models3 "BD-v2/internal/app/users/models"
)

type PostRelated struct {
	Post   *models2.Post  `json:"post,omitempty"`
	Author *models3.User  `json:"author,omitempty"`
	Thread *models.Thread `json:"thread,omitempty"`
	Forum  *models4.Forum `json:"forum,omitempty"`
}
