package entity

import (
	"github.com/CCLooMi/sql-mak/mysql/entity"
	"time"
)

type Menu struct {
	entity.IdEntity
	Name     *string `orm:"varchar(64) comment '名称'" column:"name" json:"name"`
	Href     *string `orm:"varchar(256) comment '地址'" column:"href" json:"href"`
	Pid      *string `orm:"varchar(32) comment '上级权限ID'" column:"pid" json:"pid"`
	Icon     *string `orm:"longtext comment '图标'" column:"icon" json:"icon"`
	Type     *string `orm:"varchar(16) comment '菜单类型'" column:"type" json:"type"`
	RootId   *string `orm:"varchar(32) comment '根菜单ID'" column:"rootId" json:"rootId"`
	Idx      *int    `orm:"int comment '层级深度'" column:"idx" json:"idx"`
	Position *int    `orm:"int comment '位置'" column:"position" json:"position"`
	entity.TimeEntity
	entity.AuditEntity
}

func (*Menu) TableName() string {
	return "sys_menu"
}

type Org struct {
	entity.IdEntity
	Name        *string `orm:"varchar(255) comment '组织名称'" column:"name" json:"name"`
	Description *string `orm:"varchar(255) comment '组织描述'" column:"description" json:"description"`
	entity.TimeEntity
	entity.AuditEntity
}

func (*Org) TableName() string {
	return "sys_org"
}

type OrgUser struct {
	entity.IdEntity
	UserID *string `orm:"varchar(32) comment '用户ID'" column:"user_id" json:"userID"`
	OrgID  *string `orm:"varchar(32) comment '组织ID'" column:"org_id" json:"orgID"`
	entity.TimeEntity
	entity.AuditEntity
}

func (*OrgUser) TableName() string {
	return "sys_org_user"
}

type Role struct {
	entity.IdEntity
	Name        *string `orm:"varchar(64); not null comment '角色名称'" column:"name" json:"name"`
	Code        *string `orm:"varchar(64); not null comment '角色编码'" column:"code" json:"code"`
	Description *string `orm:"varchar(255) comment '角色描述'" column:"description" json:"description"`
	entity.TimeEntity
	entity.AuditEntity
}

func (*Role) TableName() string {
	return "sys_role"
}

type RoleMenu struct {
	entity.IdEntity
	RoleId *string `orm:"varchar(32) comment '角色ID'" column:"role_id" json:"roleId"`
	MenuId *string `orm:"varchar(32) comment '视图ID'" column:"menu_id" json:"menuId"`
	entity.TimeEntity
	entity.AuditEntity
}

func (*RoleMenu) TableName() string {
	return "sys_role_menu"
}

type RolePermission struct {
	entity.IdEntity
	RoleId       *string `orm:"varchar(32) comment '角色ID'" column:"role_id" json:"roleId"`
	PermissionId *string `orm:"varchar(40) comment '权限ID'" column:"permission_id" json:"permissionId"`
	entity.TimeEntity
	entity.AuditEntity
}

func (*RolePermission) TableName() string {
	return "sys_role_permission"
}

type RoleUser struct {
	entity.IdEntity
	UserId *string `orm:"varchar(32) comment '用户ID'" column:"user_id" json:"userId"`
	RoleId *string `orm:"varchar(32) comment '角色ID'" column:"role_id" json:"roleId"`
	entity.TimeEntity
	entity.AuditEntity
}

func (*RoleUser) TableName() string {
	return "sys_role_user"
}

type Upload struct {
	entity.Id64Entity
	FileName   *string `orm:"varchar(255) comment '文件名称'" column:"file_name" json:"fileName"`
	FileType   *string `orm:"varchar(128) comment '文件类型'" column:"file_type" json:"fileType"`
	FileSize   *int64  `orm:"bigint comment '文件大小'" column:"file_size" json:"fileSize"`
	UploadSize *int64  `orm:"bigint comment '上传大小'" column:"upload_size" json:"uploadSize"`
	DelFlag    *bool   `orm:"tinyint comment '删除标识1删除0未删除'" column:"del_flag" json:"delFlag"`
	entity.TimeEntity
}

func (*Upload) TableName() string {
	return "sys_upload"
}

type Files struct {
	entity.IdEntity
	FileId   *string    `orm:"varchar(64) comment '文件ID'" column:"file_id" json:"fileId"`
	UserId   *string    `orm:"varchar(32) comment '用户ID'" column:"user_id" json:"userId"`
	FileName *string    `orm:"varchar(255) comment '文件名称'" column:"file_name" json:"fileName"`
	FileType *string    `orm:"varchar(128) comment '文件类型'" column:"file_type" json:"fileType"`
	FileSize *int64     `orm:"bigint comment '文件大小'" column:"file_size" json:"fileSize"`
	Tags     *string    `orm:"json comment '标签'" column:"tags" json:"tags"`
	DelFlag  *bool      `orm:"tinyint comment '删除标识1删除0未删除'" column:"del_flag" json:"delFlag"`
	FlagId   *string    `orm:"varchar(32) comment '标记ID'" column:"flag_id" json:"flagId"`
	FlagExp  *time.Time `orm:"datetime(6) comment '标记过期时间'" column:"flag_exp" json:"flagExp"`
	entity.TimeEntity
}

func (*Files) TableName() string {
	return "sys_files"
}

type User struct {
	entity.IdEntity
	Username string `orm:"varchar(64); not null comment '用户名'" column:"username" json:"username"`
	Nickname string `orm:"varchar(64); not null comment '用户昵称'" column:"nickname" json:"nickname"`
	Password string `orm:"varchar(64); not null comment '用户密码'" column:"password" json:"password"`
	Seed     []byte `orm:"binary(8); not null comment '密码种子'" column:"seed" json:"seed"`
	entity.TimeEntity
	entity.AuditEntity
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
	Name      *string `orm:"varchar(64); not null comment '配置名称'" column:"name" json:"name"`
	Value     *string `orm:"longtext comment '配置值'" column:"value" json:"value"`
	ValueType *string `orm:"varchar(128) comment '值类型'" column:"value_type" json:"valueType"`
	entity.TimeEntity
	entity.AuditEntity
}

func (*Config) TableName() string {
	return "sys_config"
}

type Session struct {
	entity.IdEntity
	Data    *string `orm:"longtext comment '数据'" column:"data" json:"data"`
	Expires *int64  `orm:"bigint comment '过期时间'" column:"expires" json:"expires"`
	entity.TimeEntity
}

func (*Session) TableName() string {
	return "sys_session"
}
