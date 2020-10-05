package linkcache

import (
	"encoding/json"

	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/mailbox"
)

const (
	// ServiceLinkNum topic service link num
	ServiceLinkNum = "topic_service_link_num"
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
		ID:   "",
		Body: byt,
	}
}

// DecodeLinkNumMsg decode linknum msg
func DecodeLinkNumMsg(msg *mailbox.Message) LinkNumMsg {
	lnmsg := LinkNumMsg{}
	json.Unmarshal(msg.Body, &lnmsg)
	return lnmsg
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
