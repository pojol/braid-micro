package linkcache

import (
	"encoding/json"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/mailbox"
)

const (
	// ServiceLinkNum topic service link num
	ServiceLinkNum = "braid_topic_service_link_num"

	// TopicUnlink unlink token topic
	TopicUnlink = "braid_topic_token_unlink"
	// TopicDown service down
	TopicDown = "braid_topic_service_down"
)

// LinkNumMsg msg struct
type LinkNumMsg struct {
	ID  string
	Num int
}

// EncodeLinkNumMsg encode linknum msg
func EncodeLinkNumMsg(id string, num int) *mailbox.Message {
	byt, _ := json.Marshal(&LinkNumMsg{
		ID:  id,
		Num: num,
	})

	return &mailbox.Message{
		Body: byt,
	}
}

// DecodeLinkNumMsg decode linknum msg
func DecodeLinkNumMsg(msg *mailbox.Message) LinkNumMsg {
	lnmsg := LinkNumMsg{}
	json.Unmarshal(msg.Body, &lnmsg)
	return lnmsg
}

// DownMsg down msg
type DownMsg struct {
	ID      string
	Service string
	Addr    string
}

// EncodeDownMsg encode down msg
func EncodeDownMsg(id string, service string, addr string) *mailbox.Message {
	byt, _ := json.Marshal(&DownMsg{
		ID:      id,
		Service: service,
		Addr:    addr,
	})

	return &mailbox.Message{
		Body: byt,
	}
}

// DecodeDownMsg decode down msg
func DecodeDownMsg(msg *mailbox.Message) DownMsg {
	dmsg := DownMsg{}
	json.Unmarshal(msg.Body, &dmsg)
	return dmsg
}

// ILinkCache The connector is a service that maintains the link relationship between multiple processes and users.
//
// +---parent----------+
// |                   |
// |    +--child----+  |
// |    |           |  |
// |    | token ... |  |
// |    |           |  |
// |    +-----------+  |
// |                   |
// +-------------------+
type ILinkCache interface {
	module.IModule

	// Look for existing links from the cache
	Target(token string, serviceName string) (targetAddr string, err error)

	// 将token绑定到nod
	Link(token string, target discover.Node) error

	// unlink token
	Unlink(token string) error

	// clean up the service
	Down(target discover.Node) error
}
