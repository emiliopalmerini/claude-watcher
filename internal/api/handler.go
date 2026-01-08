package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"claude-watcher/internal/database/sqlc"
)

type Handler struct {
	queries *sqlc.Queries
}

func NewHandler(queries *sqlc.Queries) *Handler {
	return &Handler{queries: queries}
}

func (h *Handler) GetChartData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	hours := rangeToHours(r.URL.Query().Get("range"))

	data := ChartData{Range: r.URL.Query().Get("range")}
	if data.Range == "" {
		data.Range = "24h"
	}

	var err error
	if data.TimeSeries, err = h.fetchTimeSeries(ctx, hours); err != nil {
		log.Printf("error fetching time series: %v", err)
	}
	if data.Models, err = h.fetchModels(ctx, hours); err != nil {
		log.Printf("error fetching model distribution: %v", err)
	}
	if data.HourOfDay, err = h.fetchHourOfDay(ctx, hours); err != nil {
		log.Printf("error fetching hour of day distribution: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) fetchTimeSeries(ctx context.Context, hours int) ([]TimePoint, error) {
	days := hours / 24
	if days < 1 {
		days = 1
	}
	rows, err := h.queries.GetDailyMetrics(ctx, sqlParam(-days))
	if err != nil {
		return nil, err
	}
	return mapSlice(rows, func(r sqlc.GetDailyMetricsRow) TimePoint {
		return TimePoint{
			Period: asString(r.Period), Sessions: r.Sessions, Cost: asFloat(r.Cost),
			Tokens: Tokens{Input: asInt(r.InputTokens), Output: asInt(r.OutputTokens), Thinking: asInt(r.ThinkingTokens)},
		}
	}), nil
}

func (h *Handler) fetchModels(ctx context.Context, hours int) ([]ModelPoint, error) {
	rows, err := h.queries.GetModelDistribution(ctx, sqlParam(-hours))
	if err != nil {
		return nil, err
	}
	return mapSlice(rows, func(r sqlc.GetModelDistributionRow) ModelPoint {
		return ModelPoint{Model: r.Model, Sessions: r.Sessions, Cost: asFloat(r.Cost)}
	}), nil
}

func (h *Handler) fetchHourOfDay(ctx context.Context, hours int) ([]HourPoint, error) {
	rows, err := h.queries.GetHourOfDayDistribution(ctx, sqlParam(-hours))
	if err != nil {
		return nil, err
	}
	return mapSlice(rows, func(r sqlc.GetHourOfDayDistributionRow) HourPoint {
		return HourPoint{Hour: r.Hour, Sessions: r.Sessions, Cost: asFloat(r.Cost)}
	}), nil
}

func sqlParam(v int) sql.NullString {
	return sql.NullString{String: strconv.Itoa(v), Valid: true}
}

func rangeToHours(r string) int {
	m := map[string]int{"7d": 168, "30d": 720, "90d": 2160}
	if h, ok := m[r]; ok {
		return h
	}
	return 168
}

func mapSlice[T, U any](s []T, f func(T) U) []U {
	r := make([]U, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}

func asFloat(v interface{}) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int64:
		return float64(x)
	}
	return 0
}

func asInt(v interface{}) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case float64:
		return int64(x)
	}
	return 0
}

func asString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	case time.Time:
		return x.Format("2006-01-02")
	}
	return ""
}
