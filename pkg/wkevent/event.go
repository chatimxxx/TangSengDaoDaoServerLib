package wkevent

import (
	"gorm.io/gorm"
)

type Type int

const (
	// None 无
	None Type = iota
	// Message 发送消息事件
	Message
	// CMD CMD
	CMD
)

func (t Type) Int() int {
	return int(t)
}

type Status int

const (
	Wait    Type = iota // 等待发布
	Success             // 发布重构
	Fail
)

func (s Status) Int() int {
	return int(s)
}

type Data struct {
	Event string      // 事件标示
	Type  Type        // 事件类型
	Data  interface{} // 事件数据
}
type Event interface {
	// 开启事件
	Begin(data *Data, tx *gorm.DB) (int64, error)
	// 提交事件
	Commit(eventId int64)
}
