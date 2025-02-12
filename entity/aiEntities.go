package entity

import (
	"github.com/CCLooMi/sql-mak/mysql/entity"
	"time"
)

type AiAssistant struct {
	entity.IdEntity
	ScriptId    *string    `orm:"varchar(32) comment '脚本ID'" column:"scriptId" json:"scriptId"`
	ScriptDesc  *string    `orm:"varchar(255) comment '脚本描述'" column:"scriptDesc" json:"scriptDesc"`
	Name        *string    `orm:"varchar(255) comment '名称'" column:"name" json:"name"`
	Conf        *string    `orm:"json comment '配置'" column:"conf" json:"conf"`
	Prompt      *string    `orm:"text comment '提示语'" column:"prompt" json:"prompt"`
	BootType    *int       `orm:"int(11) comment '1:开机启动 0:手动'" column:"bootType" json:"bootType"`
	Status      *string    `orm:"varchar(32) comment '运行状态running,stopped'" column:"status" json:"status"`
	MaxInstance *int       `orm:"int(11) comment '最大实例数'" column:"maxInstance" json:"maxInstance"`
	FlagId      *string    `orm:"varchar(32) comment '标签ID'" column:"flagId" json:"flagId"`
	FlagExp     *time.Time `orm:"datetime(6) comment '过期时间'" column:"flagExp" json:"flagExp"`
	entity.TimeEntity
}

func (*AiAssistant) TableName() string {
	return "ai_assistant"
}

type AiChatHistory struct {
	entity.IdEntity
	AssistantId *string `orm:"varchar(32) comment '助手ID'" column:"assistantId" json:"assistantId"`
	SessionId   *string `orm:"varchar(255) comment '会话ID'" column:"sessionId" json:"sessionId"`
	MsgId       *string `orm:"varchar(255) comment '消息ID'" column:"msgId" json:"msgId"`
	Subject     *string `orm:"varchar(255) comment '主题'" column:"subject" json:"subject"`
	Role        *string `orm:"varchar(32) comment '角色'" column:"role" json:"role"`
	Content     *string `orm:"longtext comment '消息内容'" column:"content" json:"content"`
	ReplyStatus *string `orm:"varchar(255) comment '回复状态'" column:"replyStatus" json:"replyStatus"`
	entity.TimeEntity
}

func (*AiChatHistory) TableName() string {
	return "ai_chat_history"
}
