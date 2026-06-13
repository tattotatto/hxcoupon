package request

type CreateStoreRequest struct {
	Name         string `json:"name" validate:"required,max=128"`
	AppID        string `json:"app_id" validate:"max=64"`
	Type         string `json:"type" validate:"required,oneof=miniprogram h5"`
	ContactName  string `json:"contact_name" validate:"max=64"`
	ContactPhone string `json:"contact_phone" validate:"max=20"`
	Remark       string `json:"remark" validate:"max=512"`
	MpAppID      string `json:"mp_appid" validate:"max=64"`
	MpPagePath   string `json:"mp_page_path" validate:"max=256"`
}

type UpdateStoreRequest struct {
	Name         string `json:"name" validate:"required,max=128"`
	ContactName  string `json:"contact_name" validate:"max=64"`
	ContactPhone string `json:"contact_phone" validate:"max=20"`
	Remark       string `json:"remark" validate:"max=512"`
	MpAppID      string `json:"mp_appid" validate:"max=64"`
	MpPagePath   string `json:"mp_page_path" validate:"max=256"`
}

type UpdateStoreStatusRequest struct {
	Status int8 `json:"status" validate:"oneof=0 1"`
}
