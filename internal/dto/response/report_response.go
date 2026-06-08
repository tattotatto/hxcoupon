package response

// TrendResponse wraps trend data
type TrendResponse struct {
	Items []TrendItem `json:"items"`
}

// StoreStatResponse for per-store statistics
type StoreStatResponse struct {
	StoreID     uint64 `json:"store_id"`
	StoreName   string `json:"store_name"`
	TotalIssued int64  `json:"total_issued"`
	TotalUsed   int64  `json:"total_used"`
}

// TemplateStatResponse for per-template statistics
type TemplateStatResponse struct {
	TemplateID   uint64  `json:"template_id"`
	TemplateName string  `json:"template_name"`
	TotalIssued  int64   `json:"total_issued"`
	TotalUsed    int64   `json:"total_used"`
	UsageRate    float64 `json:"usage_rate"`
}
