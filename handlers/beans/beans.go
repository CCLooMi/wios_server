package beans

import "wios_server/entity"

type MenuWithChecked struct {
	entity.Menu
	Checked string `json:"checked"`
}

func (*MenuWithChecked) TableName() string {
	return "MenuWithChecked"
}

type PageInfo struct {
	PageSize int                    `json:"pageSize"`
	Page     int                    `json:"page"`
	Opts     map[string]interface{} `json:"opts"`
}
