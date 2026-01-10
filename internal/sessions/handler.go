package sessions

import (
	"net/http"
	"strconv"

	"claude-watcher/internal/database/sqlc"
	apperrors "claude-watcher/internal/shared/errors"
	"claude-watcher/internal/shared/middleware"
)

type Handler struct {
	repo     Repository
	pageSize int64
}

func NewHandler(repo Repository, pageSize int64) *Handler {
	return &Handler{repo: repo, pageSize: pageSize}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	offset := int64(page-1) * h.pageSize

	sessions, err := h.repo.ListSessions(ctx, sqlc.ListSessionsParams{
		Limit:  h.pageSize,
		Offset: offset,
	})
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	count, err := h.repo.CountSessions(ctx)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}
	totalPages := int((count + h.pageSize - 1) / h.pageSize)

	data := SessionsData{
		Sessions:   sessions,
		Page:       page,
		TotalPages: totalPages,
	}

	if middleware.IsHTMX(r) {
		SessionsTable(data).Render(ctx, w)
		return
	}

	SessionsList(data).Render(ctx, w)
}
