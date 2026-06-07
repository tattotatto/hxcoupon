package request

type CreateStoreRequest struct {
	Name         string `json:"name" validate:"required,max=128"`
	Code         string `json:"code" validate:"required,len=5,alphanum"`
	AppID        string `json:"app_id" validate:"required,max=64"`
	Type         string `json:"type" validate:"required,oneof=miniprogram h5"`
	ContactName  string `json:"contact_name" validate:"max=64"`
	ContactPhone string `json:"contact_phone" validate:"max=20"`
	Remark       string `json:"remark" validate:"max=512"`
}

type UpdateStoreRequest struct {
	Name         string `json:"name" validate:"required,max=128"`
	ContactName  string `json:"contact_name" validate:"max=64"`
	ContactPhone string `json:"contact_phone" validate:"max=20"`
	Remark       string `json:"remark" validate:"max=512"`
}

type UpdateStoreStatusRequest struct {
	Status int8 `json:"status" validate:"oneof=0 1"`
}
