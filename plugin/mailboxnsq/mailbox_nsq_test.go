package mailboxnsq

import (
	"testing"

	"github.com/pojol/braid/module/mailbox"
	"github.com/stretchr/testify/assert"
)

func TestClusterMailboxParm(t *testing.T) {
	b := mailbox.GetBuilder(Name)
	b.AddOption(WithChannel("parm"))
	b.AddOption(WithLookupAddr([]string{"127.0.0.1"}))
	b.AddOption(WithNsqdAddr([]string{"127.0.0.2"}))

	mb, _ := b.Build("cluster")
	cm := mb.(*nsqMailbox)

	assert.Equal(t, cm.parm.Channel, "parm")
	assert.Equal(t, cm.parm.LookupAddres, []string{"127.0.0.1"})
	assert.Equal(t, cm.parm.Addres, []string{"127.0.0.2"})
}
