package response

type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginatedData struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Items    interface{} `json:"items"`
}

func Success(data interface{}) *Envelope {
	return &Envelope{Code: 0, Message: "success", Data: data}
}

func Error(code int, msg string) *Envelope {
	return &Envelope{Code: code, Message: msg}
}
