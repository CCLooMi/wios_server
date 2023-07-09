package entity

import (
	"time"

	"github.com/CCLooMi/sql-mak/mysql/entity"
	"github.com/shopspring/decimal"
)

type Account struct {
	entity.IdEntity
	entity.TimeEntity
	UserID  []byte          `orm:"type:binary(16); comment:'用户ID'" column:"user_id"`
	Balance decimal.Decimal `orm:"type:decimal(19,2); comment:'资金'" column:"balance"`
}

func (*Account) TableName() string {
	return "accounts"
}

type Category struct {
	entity.IdEntity
	entity.TimeEntity
	Name        string `orm:"type:varchar(50); comment:'分类名称'" column:"name"`
	Description string `orm:"type:varchar(255); comment:'分类描述'" column:"description"`
	Order       int    `orm:"comment:'分类排序'"`
}

func (*Category) TableName() string {
	return "categories"
}

type Comment struct {
	entity.IdEntity
	entity.TimeEntity
	Content  string `orm:"type:text; comment:'评论内容'" column:"content"`
	Rating   int    `orm:"comment:'评分'" column:"rating"`
	UserID   []byte `orm:"type:binary(16); comment:'用户ID'" column:"user_id"`
	TargetID []byte `orm:"type:binary(16); comment:'目标ID'" column:"target_id"`
	RootID   []byte `orm:"type:binary(16); comment:'根ID'" column:"root_id"`
}

func (*Comment) TableName() string {
	return "comments"
}

type Organization struct {
	entity.IdEntity
	entity.TimeEntity
	Name        string `orm:"type:varchar(255); not null; comment:'组织名称'" column:"name"`
	Description string `orm:"type:varchar(255); comment:'组织描述'" column:"description"`
}

func (*Organization) TableName() string {
	return "organizations"
}

type Permission struct {
	entity.IdEntity
	entity.TimeEntity
	Name        string `orm:"type:varchar(255); not null; comment:'权限名称'" column:"name"`
	Descriptor  string `orm:"type:varchar(255); not null; comment:'权限描述'" column:"descriptor"`
	Type        string `orm:"type:varchar(255); not null; comment:'权限类型'" column:"type"`
	Description string `orm:"type:varchar(255); comment:'权限描述'" column:"description"`
}

func (*Permission) TableName() string {
	return "permissions"
}

type PurchasedWpp struct {
	entity.IdEntity
	entity.TimeEntity
	UserID       []byte          `orm:"type:binary(16); comment:'用户ID'" column:"user_id"`
	WppID        []byte          `orm:"type:binary(16); comment:'应用ID'" column:"wpp_id"`
	Price        decimal.Decimal `orm:"type:decimal(10, 0); comment:'购买价格'" column:"price"`
	PurchaseTime time.Time       `orm:"comment:'购买时间'" column:"purchase_time"`
}

func (*PurchasedWpp) TableName() string {
	return "purchased_wpps"
}

type RolePermission struct {
	entity.IdEntity
	entity.TimeEntity
	RoleID       []byte `orm:"type:binary(16); comment:'角色ID'" column:"role_id"`
	PermissionID []byte `orm:"type:binary(16); comment:'权限ID'" column:"permission_id"`
}

func (*RolePermission) TableName() string {
	return "role_permissions"
}

type Role struct {
	entity.IdEntity
	entity.TimeEntity
	Name        string `orm:"type:varchar(255); not null; comment:'角色名称'" column:"name"`
	Description string `orm:"type:varchar(255); comment:'角色描述'" column:"description"`
}

func (*Role) TableName() string {
	return "roles"
}

type SchemaMigration struct {
	Version    int64      `orm:"primaryKey; not null; comment:'版本号'" column:"version"`
	InsertedAt *time.Time `orm:"comment:'插入时间'" column:"inserted_at"`
}

func (*SchemaMigration) TableName() string {
	return "schema_migrations"
}

type TMessage struct {
	entity.IdEntity
	entity.TimeEntity
	RoomID  string `orm:"type:varchar(255); comment:'房间ID'" column:"room_id"`
	Name    string `orm:"type:varchar(255); comment:'名称'" column:"name"`
	Message string `orm:"type:varchar(255); comment:'消息内容'" column:"message"`
}

func (*TMessage) TableName() string {
	return "t_messages"
}

type Upload struct {
	entity.IdEntity
	entity.TimeEntity
	FileID   []byte `orm:"type:varbinary(32); comment:'文件ID'" column:"file_id"`
	FileName string `orm:"type:varchar(255); comment:'文件名称'" column:"file_name"`
	FileType string `orm:"type:varchar(255); comment:'文件类型'" column:"file_type"`
	FileSize int64  `orm:"type:bigint(20); comment:'文件大小'" column:"file_size"`
	BizID    []byte `orm:"type:binary(16); comment:'业务ID'" column:"biz_id"`
	BizType  string `orm:"type:varchar(255); comment:'业务类型'" column:"biz_type"`
}

func (*Upload) TableName() string {
	return "uploads"
}

type UserOrganization struct {
	entity.IdEntity
	entity.TimeEntity
	UserID         []byte `orm:"type:binary(16); comment:'用户ID'" column:"user_id"`
	OrganizationID []byte `orm:"type:binary(16); comment:'组织ID'" column:"organization_id"`
}

func (*UserOrganization) TableName() string {
	return "user_organizations"
}

type UserRole struct {
	entity.IdEntity
	entity.TimeEntity
	UserID []byte `orm:"type:binary(16); comment:'用户ID'" column:"user_id"`
	RoleID []byte `orm:"type:binary(16); comment:'角色ID'" column:"role_id"`
}

func (*UserRole) TableName() string {
	return "user_roles"
}

type User struct {
	entity.IdEntity
	entity.TimeEntity
	Username string `orm:"type:varchar(255); comment:'用户名'" column:"username"`
	Password []byte `orm:"type:varbinary(32); comment:'用户密码'" column:"password"`
}

func (*User) TableName() string {
	return "users"
}

type WppCategory struct {
	entity.IdEntity
	entity.TimeEntity
	WppID      []byte `orm:"type:binary(16); comment:'应用ID'" column:"wpp_id"`
	CategoryID []byte `orm:"type:binary(16); comment:'分类ID'" column:"category_id"`
}

func (*WppCategory) TableName() string {
	return "wpp_categories"
}

type Wpp struct {
	entity.IdEntity
	entity.TimeEntity
	Name        string `orm:"type:varchar(64); comment:'应用名称'" column:"name"`
	Description string `orm:"type:text; comment:'描述'" column:"description"`
	Version     string `orm:"type:varchar(255); comment:'版本号'" column:"version"`
	DeveloperID []byte `orm:"type:binary(16); comment:'开发者ID'" column:"developer_id"`
	FileID      []byte `orm:"type:varbinary(32); comment:'文件ID'" column:"file_id"`
}

func (*Wpp) TableName() string {
	return "wpps"
}
