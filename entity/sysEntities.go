package entity

import (
	"github.com/CCLooMi/sql-mak/mysql/entity"
)

type Menu struct {
	entity.IdEntity
	Name     string `orm:"varchar(64) comment '名称'" column:"name" json:"name"`
	Href     string `orm:"varchar(256) comment '地址'" column:"href" json:"href"`
	Pid      string `orm:"varchar(32) comment '上级权限ID'" column:"pid" json:"pid"`
	Icon     string `orm:"longtext comment '图标'" column:"icon" json:"icon"`
	Type     string `orm:"varchar(16) comment '菜单类型'" column:"type" json:"type"`
	RootId   string `orm:"varchar(32) comment '根菜单ID'" column:"rootId" json:"rootId"`
	Idx      int    `orm:"int comment '层级深度'" column:"idx" json:"idx"`
	Position int    `orm:"int comment '位置'" column:"position" json:"position"`
	entity.TimeEntity
}

func (*Menu) TableName() string {
	return "sys_menu"
}

type Org struct {
	entity.IdEntity
	Name        string `orm:"varchar(255) comment '组织名称'" column:"name" json:"name"`
	Description string `orm:"varchar(255) comment '组织描述'" column:"description" json:"description"`
	entity.TimeEntity
}

func (*Org) TableName() string {
	return "sys_org"
}

type OrgUser struct {
	entity.IdEntity
	UserID string `orm:"varchar(32) comment '用户ID'" column:"user_id" json:"userID"`
	OrgID  string `orm:"varchar(32) comment '组织ID'" column:"org_id" json:"orgID"`
	entity.TimeEntity
}

func (*OrgUser) TableName() string {
	return "sys_org_user"
}

type Permission struct {
	entity.IdEntity
	Name        string `orm:"varchar(64) not null comment '权限名称'" column:"name" json:"name"`
	Type        string `orm:"varchar(32) not null comment '权限类型'" column:"type" json:"type"`
	Description string `orm:"varchar(255) comment '权限描述'" column:"description" json:"description"`
	entity.TimeEntity
}

func (*Permission) TableName() string {
	return "sys_permission"
}

type Role struct {
	entity.IdEntity
	Name        string `orm:"varchar(64); not null comment '角色名称'" column:"name" json:"name"`
	Code        string `orm:"varchar(64); not null comment '角色编码'" column:"code" json:"code"`
	Description string `orm:"varchar(255) comment '角色描述'" column:"description" json:"description"`
	entity.TimeEntity
}

func (*Role) TableName() string {
	return "sys_role"
}

type RoleMenu struct {
	entity.IdEntity
	RoleId string `orm:"varchar(32) comment '角色ID'" column:"role_id" json:"roleId"`
	MenuId string `orm:"varchar(32) comment '视图ID'" column:"menu_id" json:"menuId"`
	entity.TimeEntity
}

func (*RoleMenu) TableName() string {
	return "sys_role_menu"
}

type RolePermission struct {
	entity.IdEntity
	RoleId       string `orm:"varchar(32) comment '角色ID'" column:"role_id" json:"roleId"`
	PermissionId string `orm:"varchar(32) comment '权限ID'" column:"permission_id" json:"permissionId"`
	entity.TimeEntity
}

func (*RolePermission) TableName() string {
	return "sys_role_permission"
}

type RoleUser struct {
	entity.IdEntity
	UserId *string `orm:"varchar(32) comment '用户ID'" column:"user_id" json:"userId"`
	RoleId *string `orm:"varchar(32) comment '角色ID'" column:"role_id" json:"roleId"`
	entity.TimeEntity
}

func (*RoleUser) TableName() string {
	return "sys_role_user"
}

type Upload struct {
	entity.IdEntity
	FileId   string `orm:"varchar(64) comment '文件ID'" column:"file_id" json:"fileId"`
	FileName string `orm:"varchar(255) comment '文件名称'" column:"file_name" json:"fileName"`
	FileType string `orm:"varchar(64) comment '文件类型'" column:"file_type" json:"fileType"`
	FileSize int64  `orm:"bigint comment '文件大小'" column:"file_size" json:"fileSize"`
	BizId    string `orm:"varchar(32) comment '业务ID'" column:"biz_id" json:"bizId"`
	BizType  string `orm:"varchar(255) comment '业务类型'" column:"biz_type" json:"bizType"`
	entity.TimeEntity
}

func (*Upload) TableName() string {
	return "sys_upload"
}

type User struct {
	entity.IdEntity
	Username string `orm:"varchar(64); not null comment '用户名'" column:"username" json:"username"`
	Nickname string `orm:"varchar(64); not null comment '用户昵称'" column:"nickname" json:"nickname"`
	Password string `orm:"varchar(64); not null comment '用户密码'" column:"password" json:"password"`
	Seed     []byte `orm:"binary(8); not null comment '密码种子'" column:"seed" json:"seed"`
	entity.TimeEntity
}

func (*User) TableName() string {
	return "sys_user"
}

type Api struct {
	entity.IdEntity
	Desc     *string `orm:"varchar(255) comment '接口描述'" column:"desc" json:"desc"`
	Script   *string `orm:"longtext comment '接口脚本'" column:"script" json:"script"`
	Type     *string `orm:"varchar(32) comment '接口类型'" column:"type" json:"type"`
	Category *string `orm:"varchar(32) comment '接口分类'" column:"category" json:"category"`
	Status   *string `orm:"varchar(32) comment '接口状态'" column:"status" json:"status"`
	entity.TimeEntity
	entity.AuditEntity
}

func (*Api) TableName() string {
	return "sys_api"
}

type Config struct {
	entity.IdEntity
	Key      string `orm:"varchar(64); not null comment '配置key'" column:"key" json:"key"`
	Category string `orm:"varchar(64) comment '配置分类'" column:"category" json:"category"`
	Value    string `orm:"longtext comment '配置值'" column:"value" json:"value"`
	entity.TimeEntity
}

func (*Config) TableName() string {
	return "sys_config"
}
