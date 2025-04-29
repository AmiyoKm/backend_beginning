package main

import (
	"net/http"
)

func (app *Application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	//TODO : PAGINATION , FILTERS
	ctx := r.Context()
	feed, err := app.Store.Posts.GetUserFeed(ctx, int64(100))
	for i := range feed {
		comments, err := app.Store.Comments.GetByPostID(ctx, feed[i].ID)

		if err != nil {
			app.notFoundResponse(w, r, err)
			return
		}
		feed[i].Comments = comments

	}
	if err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}
}
