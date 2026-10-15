package responses

// Error describes the common API error payload.
type Error struct {
	Code      string `json:"code" example:"unauthorized"`
	Message   string `json:"message" example:"valid Bearer access token required"`
	Parameter string `json:"parameter,omitempty" example:"id"`
}

// Health describes database connectivity, not schema readiness.
type Health struct {
	Status   string `json:"status" example:"ok"`
	Database string `json:"database" example:"ok"`
	Code     string `json:"code,omitempty" example:"database_unavailable"`
}
