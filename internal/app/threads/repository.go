package threads

import (
	models2 "BD-v2/internal/app/posts/models"
	models3 "BD-v2/internal/app/posts/post_related"
	"BD-v2/internal/app/threads/models"
	"context"
)

type Repository interface {
	CreateThread(ctx context.Context, thread *models.Thread) (*models.Thread, error)
	FindThreadSlug(ctx context.Context, slug string) (*models.Thread, error)
	FindThreadID(ctx context.Context, threadID int) (*models.Thread, error)
	CreatePost(post []*models2.Post, thread *models.Thread) ([]*models2.Post, error)
	GetPosts(threadID, limit, since int, desc bool) ([]*models2.Post, error)
	GetPostsTree(threadID, limit, since int, desc bool) ([]*models2.Post, error)
	GetPostsParentTree(threadID, limit, since int, desc bool) ([]*models2.Post, error)
	AddVoice(ctx context.Context, voice *models.Vote) error
	GetVoice(ctx context.Context, voice *models.Vote) (*models.Vote, error)
	UpdateVoice(ctx context.Context, voice *models.Vote) (int, error)
	UpdateThreadID(ctx context.Context, thread *models.Thread) error
	UpdateThreadSlug(ctx context.Context, thread *models.Thread) error
	GetOnePost(ctx context.Context, postID int, rel []string) (*models3.PostRelated, error)
	GetPostByID(ctx context.Context, postID int) (*models2.Post, error)
	UpdatePost(ctx context.Context, post *models2.Post) error
}
