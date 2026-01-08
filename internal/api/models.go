package api

// ChartData aggregates all chart data for the dashboard
type ChartData struct {
	TimeSeries []TimePoint  `json:"timeSeries"`
	Models     []ModelPoint `json:"models"`
	HourOfDay  []HourPoint  `json:"hourOfDay"`
	Range      string       `json:"range"`
}

type TimePoint struct {
	Period   string  `json:"period"`
	Sessions int64   `json:"sessions"`
	Cost     float64 `json:"cost"`
	Tokens   Tokens  `json:"tokens"`
}

type Tokens struct {
	Input    int64 `json:"input"`
	Output   int64 `json:"output"`
	Thinking int64 `json:"thinking"`
}

type ModelPoint struct {
	Model    string  `json:"model"`
	Sessions int64   `json:"sessions"`
	Cost     float64 `json:"cost"`
}

type HourPoint struct {
	Hour     int64   `json:"hour"`
	Sessions int64   `json:"sessions"`
	Cost     float64 `json:"cost"`
}
