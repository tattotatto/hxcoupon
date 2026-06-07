package response

import "hxcoupon/internal/model"

type StoreResponse struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	AppID        string `json:"app_id"`
	Type         string `json:"type"`
	Status       int8   `json:"status"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	Remark       string `json:"remark"`
}

func ToStoreResponse(s *model.Store) *StoreResponse {
	return &StoreResponse{
		ID:           s.ID,
		Name:         s.Name,
		Code:         s.Code,
		AppID:        s.AppID,
		Type:         s.Type,
		Status:       s.Status,
		ContactName:  s.ContactName,
		ContactPhone: s.ContactPhone,
		Remark:       s.Remark,
	}
}

type StoreWithCredentialsResponse struct {
	StoreResponse
	Credentials *CredentialResponse `json:"credentials,omitempty"`
}

type CredentialResponse struct {
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
}
