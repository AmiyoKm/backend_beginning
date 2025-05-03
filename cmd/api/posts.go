package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/AmiyoKm/go-backend/internal/store"
	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}
// CreatePost godoc
//
//	@Summary		Creates a post
//	@Description	Creates a post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreatePostPayload	true	"Post payload"
//	@Success		201		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
func (app *Application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload

	if err := readJson(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := getUserFromContext(r)

	userId := user.ID
	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		UserID:  int64(userId),
		Tags:    payload.Tags,
	}
	ctx := r.Context()

	if err := app.Store.Posts.Create(ctx, post); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return

	}
}
// GetPost godoc
//
//	@Summary		Fetches a post
//	@Description	Fetches a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	store.Post
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [get]
func (app *Application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	ctx := r.Context()
	comments, err := app.Store.Comments.GetByPostID(ctx, post.ID)

	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	post.Comments = comments
	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}
// DeletePost godoc
//
//	@Summary		Deletes a post
//	@Description	Delete a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		204	{object}	string
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [delete]
func (app *Application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "postID")
	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	ctx := r.Context()

	err = app.Store.Posts.Delete(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrorNotFound):
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerErrorResponse(w, r, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}
// UpdatePost godoc
//
//	@Summary		Updates a post
//	@Description	Updates a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Post ID"
//	@Param			payload	body		UpdatePostPayload	true	"Post payload"
//	@Success		200		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [patch]
func (app *Application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	ctx := r.Context()
	var payload UpdatePostPayload
	if err := readJson(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := validate.Struct(payload); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if err := app.Store.Posts.Update(ctx, post); err != nil {
		switch {
		case errors.Is(err, store.ErrorNotFound):
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerErrorResponse(w, r, err)
			return
		}
	}
	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

}

type CommentPayload struct {
	Content string `json:"content" validate:"required,max=400"`
	UserID  int    `json:"user_id" validate:"required"`
}
// CreateComment godoc
//
//	@Summary		Create a comment
//	@Description	Creates a comment on a post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int				true	"Post ID"
//	@Param			payload	body		CommentPayload	true	"Comment payload"
//	@Success		200		{object}	store.Comment
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id}/comments [post]
func (app *Application) createCommentsHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	ctx := r.Context()

	var payload CommentPayload

	if err := readJson(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	comment := &store.Comment{
		PostID:  post.ID,
		UserID:  int64(payload.UserID),
		Content: payload.Content,
	}
	if err := app.Store.Comments.Create(ctx, comment); err != nil {
		switch {
		case errors.Is(err, store.ErrorNotFound):
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerErrorResponse(w, r, err)
			return
		}
	}
	if err := app.jsonResponse(w, http.StatusCreated, comment); err != nil {
		app.internalServerErrorResponse(w, r, err)
	}

}

func (app *Application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "postID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerErrorResponse(w, r, err)
			return
		}
		ctx := r.Context()
		post, err := app.Store.Posts.GetByID(ctx, id)

		if err != nil {
			switch {
			case errors.Is(err, store.ErrorNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerErrorResponse(w, r, err)
			}
			return
		}
		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
