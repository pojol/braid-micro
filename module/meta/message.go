package meta

import (
	"time"
)

// Message 消息体
type Message struct {
	Body []byte

	id        string
	timestamp int64
}

func (msg *Message) ID() string {
	return msg.id
}

func (msg *Message) Timestamp() int64 {
	return msg.timestamp
}

func CreateMessage(id string, body []byte) *Message {

	return &Message{
		id:        id,
		timestamp: time.Now().UnixMilli(),
		Body:      body,
	}

}
