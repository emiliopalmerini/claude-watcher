package web

import (
	"net/http"

	"github.com/emiliopalmerini/mclaude/internal/web/templates"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	pricing, _ := queries.ListModelPricing(ctx)

	models := make([]templates.ModelPricing, 0, len(pricing))
	for _, p := range pricing {
		model := templates.ModelPricing{
			ID:               p.ID,
			DisplayName:      p.DisplayName,
			InputPerMillion:  p.InputPerMillion,
			OutputPerMillion: p.OutputPerMillion,
			IsDefault:        p.IsDefault == 1,
		}
		if p.CacheReadPerMillion.Valid {
			model.CacheReadPerMillion = p.CacheReadPerMillion.Float64
		}
		if p.CacheWritePerMillion.Valid {
			model.CacheWritePerMillion = p.CacheWritePerMillion.Float64
		}
		models = append(models, model)
	}

	templates.Settings(models).Render(ctx, w)
}
