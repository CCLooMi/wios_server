package entity

import (
	"github.com/CCLooMi/sql-mak/mysql/entity"
	"time"
)

type Account struct {
	entity.IdEntity
	UserId  *string  `orm:"varchar(32) comment '用户ID'" column:"user_id" json:"userId"`
	Balance *float64 `orm:"decimal(19,2) comment '资金'" column:"balance" json:"balance"`
	entity.TimeEntity
}

func (*Account) TableName() string {
	return "t_account"
}

type Category struct {
	entity.IdEntity
	Name        *string `orm:"varchar(64) comment '分类名称'" column:"name" json:"name"`
	Description *string `orm:"varchar(255) comment '分类描述'" column:"description" json:"description"`
	Order       *int    `orm:"int; default:0 comment '分类排序'" column:"order" json:"order"`
	entity.TimeEntity
}

func (*Category) TableName() string {
	return "t_category"
}

type Comment struct {
	entity.IdEntity
	Content  *string `orm:"text comment '评论内容'" column:"content" json:"content"`
	Rating   *int    `orm:"int comment '评分'" column:"rating" json:"rating"`
	UserId   *string `orm:"varchar(32) comment '用户ID'" column:"user_id" json:"userId"`
	TargetId *string `orm:"varchar(32) comment '目标ID'" column:"target_id" json:"targetId"`
	RootId   *string `orm:"varchar(32) comment '根ID'" column:"root_id" json:"rootId"`
	entity.TimeEntity
}

func (*Comment) TableName() string {
	return "t_comment"
}

type PurchasedWpp struct {
	entity.IdEntity
	UserId       *string   `orm:"varchar(32) comment '用户ID'" column:"user_id" json:"userId"`
	WppId        *string   `orm:"varchar(32) comment '应用ID'" column:"wpp_id" json:"wppId"`
	Price        *int64    `orm:"decimal(10,0) comment '购买价格'" column:"price" json:"price"`
	PurchaseTime time.Time `orm:"datetime comment '购买时间'" column:"purchase_time" json:"purchaseTime"`
	entity.TimeEntity
}

func (*PurchasedWpp) TableName() string {
	return "t_purchased_wpp"
}

type ReleaseNote struct {
	entity.IdEntity
	WppId       *string `orm:"varchar(32) comment '应用ID'" column:"wpp_id" json:"wppId"`
	Version     *string `orm:"varchar(32) comment '版本号'" column:"version" json:"version"`
	ReleaseNote *string `orm:"varchar(255) comment '发布日志'" column:"release_note" json:"releaseNote"`
	DeveloperId *string `orm:"varchar(32) comment '开发者ID'" column:"developer_id" json:"developerId"`
	FileId      *string `orm:"varchar(64) comment '文件ID'" column:"file_id" json:"fileId"`
	entity.TimeEntity
}

func (*ReleaseNote) TableName() string {
	return "t_wpp_release_note"
}

type Wpp struct {
	entity.IdEntity
	Name          *string `orm:"varchar(64) comment '应用名称'" column:"name" json:"name"`
	Manifest      *string `orm:"longtext comment '元数据'" column:"manifest" json:"manifest"`
	LatestVersion *string `orm:"varchar(32) comment '最新版本号'" column:"latest_version" json:"LatestVersion"`
	DeveloperId   *string `orm:"varchar(32) comment '开发者ID'" column:"developer_id" json:"developerId"`
	FileId        *string `orm:"varchar(64) comment '文件ID'" column:"file_id" json:"fileId"`
	entity.TimeEntity
}

func (*Wpp) TableName() string {
	return "t_wpp"
}

type StoreUser struct {
	entity.IdEntity
	Username *string `orm:"varchar(64); comment '用户名'" column:"username" json:"username"`
	Nickname *string `orm:"varchar(64); comment '用户昵称'" column:"nickname" json:"nickname"`
	Email    *string `orm:"varchar(256) comment '邮箱'" column:"email" json:"email"`
	Phone    *string `orm:"varchar(16) comment '手机号'" column:"phone" json:"phone"`
	Avatar   *string `orm:"longtext comment '头像'" column:"avatar" json:"avatar"`
	Password string  `orm:"varchar(64); not null comment '用户密码'" column:"password" json:"password"`
	Seed     []byte  `orm:"binary(8); not null comment '密码种子'" column:"seed" json:"seed"`
	entity.TimeEntity
}

func (*StoreUser) TableName() string {
	return "t_store_user"
}
