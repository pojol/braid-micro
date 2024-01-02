package components

import (
	"testing"

	"github.com/pojol/braid-go/components/depends/blog"
)

func TestBuild(t *testing.T) {

	d := DefaultDirector{
		Opts: &DirectorOpts{
			LogOpts: []blog.Option{
				blog.WithLevel(int(blog.DebugLevel)),
			},
		},
	}

	d.Build()

}
