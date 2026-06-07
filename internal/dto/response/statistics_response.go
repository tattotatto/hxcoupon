package response

type OverviewResponse struct {
	TotalStores    int64   `json:"total_stores"`
	TotalTemplates int64   `json:"total_templates"`
	TotalIssued    int64   `json:"total_issued"`
	TotalUsed      int64   `json:"total_used"`
	UsageRate      float64 `json:"usage_rate"`
	TodayIssued    int64   `json:"today_issued"`
	TodayUsed      int64   `json:"today_used"`
}

type TrendItem struct {
	Date   string `json:"date"`
	Issued int64  `json:"issued"`
	Used   int64  `json:"used"`
}
