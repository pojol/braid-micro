package braid

import (
	"errors"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/linkcache"
)

func (b *Braid) Discover(opts ...discover.Option) module.IModule {

	d := discover.Build(b.name, opts...)
	b.discoverPtr = d

	return d
}

func (b *Braid) LinkCache(opts ...linkcache.Option) module.IModule {

	if b.pubsub == nil {
		panic(errors.New("LinkCache module need depend Pubsub"))
	}

	if b.redis == nil {
		panic(errors.New("LinkCache module need depend Redis"))
	}

	l := linkcache.Build(b.name, b.pubsub, b.redis, opts...)
	b.linkcachePtr = l

	return l
}

func (b *Braid) Elector(opts ...elector.Option) module.IModule {
	if b.pubsub == nil {
		panic(errors.New("LinkCache module need depend Pubsub"))
	}

	e := elector.Build(b.name, b.pubsub, opts...)
	b.electorPtr = e

	return e
}
