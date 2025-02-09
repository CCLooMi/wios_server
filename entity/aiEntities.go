package entity

import "github.com/CCLooMi/sql-mak/mysql/entity"

type AiAssistant struct {
	entity.IdEntity
	Name   *string `orm:"varchar(255) comment '名称'" column:"name" json:"name"`
	Conf   *string `orm:"json comment '配置'" column:"conf" json:"conf"`
	Prompt *string `orm:"text comment '提示语'" column:"prompt" json:"prompt"`
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
	entity.TimeEntity
}

func (*AiChatHistory) TableName() string {
	return "ai_chat_history"
}
