package live

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/live", h.Show)
	r.Get("/live/sessions", h.SessionsPartial)
}
