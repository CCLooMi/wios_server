package entity

import (
	"github.com/CCLooMi/sql-mak/mysql/entity"
)

type Menu struct {
	entity.IdEntity
	Name     string `orm:"varchar(64) comment '名称'" column:"name"`
	URL      string `orm:"varchar(256) comment '地址'" column:"url"`
	Pid      string `orm:"varchar(32) comment '上级权限ID'" column:"pid"`
	Icon     string `orm:"longtext comment '图标'" column:"icon"`
	Type     string `orm:"varchar(16) comment '菜单类型'" column:"type"`
	RootId   string `orm:"varchar(32) comment '根菜单ID'" column:"rootId"`
	Idx      int    `orm:"int comment '层级深度'" column:"idx"`
	Position int    `orm:"int comment '位置'" column:"position"`
	entity.TimeEntity
}

func (*Menu) TableName() string {
	return "sys_menu"
}

type Org struct {
	entity.IdEntity
	Name        string `orm:"varchar(255) comment '组织名称'" column:"name"`
	Description string `orm:"varchar(255) comment '组织描述'" column:"description"`
	entity.TimeEntity
}

func (*Org) TableName() string {
	return "sys_org"
}

type OrgUser struct {
	entity.IdEntity
	UserID string `orm:"varchar(32) comment '用户ID'" column:"user_id"`
	OrgID  string `orm:"varchar(32) comment '组织ID'" column:"org_id"`
	entity.TimeEntity
}

func (*OrgUser) TableName() string {
	return "sys_org_user"
}

type Permission struct {
	entity.IdEntity
	Name        string `orm:"varchar(64) not null comment '权限名称'" column:"name"`
	Type        string `orm:"varchar(32) not null comment '权限类型'" column:"type"`
	Description string `orm:"varchar(255) comment '权限描述'" column:"description"`
	entity.TimeEntity
}

func (*Permission) TableName() string {
	return "sys_permission"
}

type Role struct {
	entity.IdEntity
	Name        string `orm:"varchar(64); not null comment '角色名称'" column:"name"`
	Description string `orm:"varchar(255) comment '角色描述'" column:"description"`
	entity.TimeEntity
}

func (*Role) TableName() string {
	return "sys_role"
}

type RoleMenu struct {
	entity.IdEntity
	RoleId string `orm:"varchar(32) comment '角色ID'" column:"role_id"`
	MenuId string `orm:"varchar(32) comment '视图ID'" column:"menu_id"`
	entity.TimeEntity
}

func (*RoleMenu) TableName() string {
	return "sys_role_menu"
}

type RolePermission struct {
	entity.IdEntity
	RoleId       string `orm:"varchar(32) comment '角色ID'" column:"role_id"`
	PermissionId string `orm:"varchar(32) comment '权限ID'" column:"permission_id"`
	entity.TimeEntity
}

func (*RolePermission) TableName() string {
	return "sys_role_permission"
}

type RoleUser struct {
	entity.IdEntity
	UserId string `orm:"varchar(32) comment '用户ID'" column:"user_id"`
	RoleId string `orm:"varchar(32) comment '角色ID'" column:"role_id"`
	entity.TimeEntity
}

func (*RoleUser) TableName() string {
	return "sys_role_user"
}

type Upload struct {
	entity.IdEntity
	FileId   string `orm:"varchar(64) comment '文件ID'" column:"file_id"`
	FileName string `orm:"varchar(255) comment '文件名称'" column:"file_name"`
	FileType string `orm:"varchar(32) comment '文件类型'" column:"file_type"`
	FileSize int64  `orm:"bigint comment '文件大小'" column:"file_size"`
	BizId    string `orm:"varchar(32) comment '业务ID'" column:"biz_id"`
	BizType  string `orm:"varchar(255) comment '业务类型'" column:"biz_type"`
	entity.TimeEntity
}

func (*Upload) TableName() string {
	return "sys_upload"
}

type User struct {
	entity.IdEntity
	Username string `orm:"varchar(64); not null comment '用户名'" column:"username"`
	Password string `orm:"varchar(64); not null comment '用户密码'" column:"password"`
	Seed     []byte `orm:"binary(8); not null comment '密码种子'" column:"seed"`
	entity.TimeEntity
}

func (*User) TableName() string {
	return "sys_user"
}

type Api struct {
	entity.IdEntity
	Desc     string `orm:"varchar(255) comment '接口描述'" column:"desc"`
	Script   string `orm:"longtext comment '接口脚本'" column:"script"`
	Type     string `orm:"varchar(32) comment '接口类型'" column:"type"`
	Category string `orm:"varchar(32) comment '接口分类'" column:"category"`
	Status   string `orm:"varchar(32) comment '接口状态'" column:"status"`
	entity.TimeEntity
	entity.AuditEntity
}

func (*Api) TableName() string {
	return "sys_api"
}
