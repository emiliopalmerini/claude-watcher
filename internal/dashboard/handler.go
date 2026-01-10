package dashboard

import (
	"net/http"

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

	metrics, err := h.repo.GetDashboardMetrics(ctx)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	today, err := h.repo.GetTodayMetrics(ctx)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	week, err := h.repo.GetWeekMetrics(ctx)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	data := DashboardData{
		Metrics: metrics,
		Today:   today,
		Week:    week,
	}

	Dashboard(data).Render(ctx, w)
}
