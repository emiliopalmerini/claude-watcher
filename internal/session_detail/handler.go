package session_detail

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	apperrors "claude-watcher/internal/shared/errors"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionID := chi.URLParam(r, "sessionID")

	session, err := h.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		apperrors.HandleError(w, apperrors.HandleDBError(err, "session not found"))
		return
	}

	SessionDetail(session).Render(ctx, w)
}
