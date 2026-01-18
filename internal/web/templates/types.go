package templates

type DashboardStats struct {
	SessionCount     int64
	TotalTokens      int64
	TotalCost        float64
	TotalTurns       int64
	TokenInput       int64
	TokenOutput      int64
	CacheRead        int64
	CacheWrite       int64
	TotalErrors      int64
	ActiveExperiment string
	TopTools         []ToolUsage
	RecentSessions   []SessionSummary
}

type ToolUsage struct {
	Name  string
	Count int64
}

type SessionSummary struct {
	ID           string
	ProjectID    string
	ExperimentID string
	CreatedAt    string
	ExitReason   string
	Turns        int64
	Tokens       int64
	Cost         float64
}

type SessionDetail struct {
	ID                    string
	ProjectID             string
	ExperimentID          string
	Cwd                   string
	PermissionMode        string
	ExitReason            string
	StartedAt             string
	EndedAt               string
	DurationSeconds       int64
	CreatedAt             string
	MessageCountUser      int64
	MessageCountAssistant int64
	TurnCount             int64
	TokenInput            int64
	TokenOutput           int64
	TokenCacheRead        int64
	TokenCacheWrite       int64
	CostEstimateUsd       float64
	ErrorCount            int64
	Tools                 []ToolUsage
	Files                 []FileOperation
}

type FileOperation struct {
	Path      string
	Operation string
	Count     int64
}

type Experiment struct {
	ID          string
	Name        string
	Description string
	Hypothesis  string
	StartedAt   string
	EndedAt     string
	IsActive    bool
	CreatedAt   string
	// Stats
	SessionCount   int64
	TotalTokens    int64
	TotalCost      float64
	TokensPerSess  int64
	CostPerSession float64
}

type ExperimentDetail struct {
	ID          string
	Name        string
	Description string
	Hypothesis  string
	StartedAt   string
	EndedAt     string
	IsActive    bool
	CreatedAt   string
	// Stats
	SessionCount      int64
	TotalTurns        int64
	UserMessages      int64
	AssistantMessages int64
	TotalErrors       int64
	TokenInput        int64
	TokenOutput       int64
	CacheRead         int64
	CacheWrite        int64
	TotalTokens       int64
	TotalCost         float64
	TokensPerSession  int64
	CostPerSession    float64
	// Top tools
	TopTools []ToolUsage
	// Recent sessions
	RecentSessions []SessionSummary
}

type ModelPricing struct {
	ID                   string
	DisplayName          string
	InputPerMillion      float64
	OutputPerMillion     float64
	CacheReadPerMillion  float64
	CacheWritePerMillion float64
	IsDefault            bool
}
