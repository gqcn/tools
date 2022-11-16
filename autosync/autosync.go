package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
)

type AutoSync struct {
	g.Meta `name:"AutoSync" root:"Sync" dc:"监听同步管理器"`
	writer *Writer
}

func main() {
	var (
		err error
		ctx = gctx.GetInitCtx()
	)
	cmd, err := gcmd.NewFromObject(&AutoSync{
		writer: NewWriter(),
	})
	if err != nil {
		g.Log().Fatalf(ctx, `%+v`, err)
	}
	cmd.Run(ctx)
}
