package entity

import (
	"github.com/CCLooMi/sql-mak/mysql/entity"
	"time"
)

type Account struct {
	entity.IdEntity
	UserId  *entity.ID `orm:"binary(16) comment '用户ID'" column:"user_id"`
	Balance float64    `orm:"decimal(19,2) comment '资金'" column:"balance"`
	entity.TimeEntity
}

func (*Account) TableName() string {
	return "t_account"
}

type Category struct {
	entity.IdEntity
	Name        string `orm:"varchar(64) comment '分类名称'" column:"name"`
	Description string `orm:"varchar(255) comment '分类描述'" column:"description"`
	Order       int    `orm:"int; default:0 comment '分类排序'" column:"order"`
	entity.TimeEntity
}

func (*Category) TableName() string {
	return "t_category"
}

type Comment struct {
	entity.IdEntity
	Content  string     `orm:"text comment '评论内容'" column:"content"`
	Rating   int        `orm:"int comment '评分'" column:"rating"`
	UserId   *entity.ID `orm:"binary(16) comment '用户ID'" column:"user_id"`
	TargetId *entity.ID `orm:"binary(16) comment '目标ID'" column:"target_id"`
	RootId   *entity.ID `orm:"binary(16) comment '根ID'" column:"root_id"`
	entity.TimeEntity
}

func (*Comment) TableName() string {
	return "t_comment"
}

type PurchasedWpp struct {
	entity.IdEntity
	UserId       *entity.ID `orm:"binary(16) comment '用户ID'" column:"user_id"`
	WppId        *entity.ID `orm:"binary(16) comment '应用ID'" column:"wpp_id"`
	Price        int64      `orm:"decimal(10,0) comment '购买价格'" column:"price"`
	PurchaseTime time.Time  `orm:"datetime comment '购买时间'" column:"purchase_time"`
	entity.TimeEntity
}

func (*PurchasedWpp) TableName() string {
	return "t_purchased_wpp"
}

type Wpp struct {
	entity.IdEntity
	Name        string     `orm:"varchar(64) comment '应用名称'" column:"name"`
	Description string     `orm:"text comment '描述'" column:"description"`
	Version     string     `orm:"varchar(32) comment '版本号'" column:"version"`
	DeveloperId *entity.ID `orm:"binary(16) comment '开发者ID'" column:"developer_id"`
	FileId      *entity.ID `orm:"varbinary(32) comment '文件ID'" column:"file_id"`
	entity.TimeEntity
}

func (*Wpp) TableName() string {
	return "t_wpps"
}
