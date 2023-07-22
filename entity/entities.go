package entity

import (
	"github.com/CCLooMi/sql-mak/mysql/entity"
	"time"
)

type Account struct {
	entity.IdEntity
	UserId  []byte  `orm:"type:binary(16); comment:'用户ID'" column:"user_id"`
	Balance float64 `orm:"type:decimal(19,2); comment:'资金'" column:"balance"`
	entity.TimeEntity
}

func (*Account) TableName() string {
	return "t_account"
}

type Category struct {
	entity.IdEntity
	Name        string `orm:"type:varchar(255); comment:'分类名称'" column:"name"`
	Description string `orm:"type:varchar(255); comment:'分类描述'" column:"description"`
	Order       int    `orm:"type:int; default:0; comment:'分类排序'" column:"order"`
	entity.TimeEntity
}

func (*Category) TableName() string {
	return "t_category"
}

type Comment struct {
	entity.IdEntity
	Content  string `orm:"type:text; comment:'评论内容'" column:"content"`
	Rating   int    `orm:"type:int; comment:'评分'" column:"rating"`
	UserId   []byte `orm:"type:binary(16); comment:'用户ID'" column:"user_id"`
	TargetId []byte `orm:"type:binary(16); comment:'目标ID'" column:"target_id"`
	RootId   []byte `orm:"type:binary(16); comment:'根ID'" column:"root_id"`
	entity.TimeEntity
}

func (*Comment) TableName() string {
	return "t_comment"
}

type PurchasedWpp struct {
	entity.IdEntity
	UserId       []byte    `orm:"type:binary(16); comment:'用户ID'" column:"user_id"`
	WppId        []byte    `orm:"type:binary(16); comment:'应用ID'" column:"wpp_id"`
	Price        int64     `orm:"type:decimal(10,0); comment:'购买价格'" column:"price"`
	PurchaseTime time.Time `orm:"type:datetime; comment:'购买时间'" column:"purchase_time"`
	entity.TimeEntity
}

func (*PurchasedWpp) TableName() string {
	return "t_purchased_wpp"
}

type Wpp struct {
	entity.IdEntity
	Name        string `orm:"type:varchar(64); comment:'应用名称'" column:"name"`
	Description string `orm:"type:text; comment:'描述'" column:"description"`
	Version     string `orm:"type:varchar(32); comment:'版本号'" column:"version"`
	DeveloperId []byte `orm:"type:binary(16); comment:'开发者ID'" column:"developer_id"`
	FileId      []byte `orm:"type:varbinary(32); comment:'文件ID'" column:"file_id"`
	entity.TimeEntity
}

func (*Wpp) TableName() string {
	return "t_wpps"
}
