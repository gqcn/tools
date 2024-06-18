package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gproc"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/gogf/gf/v2/text/gstr"
)

type SyncInput struct {
	g.Meta      `dc:"通过本地fswatch监听本地目录变化，并使用rsync自动同步到远端服务器目录"`
	LocalDir    string `v:"required" dc:"本地目录"`
	RemoteDir   string `v:"required" dc:"远端服务器目录"`
	RemoteHost  string `v:"required" dc:"远端服务器地址"`
	RemotePort  uint   `v:"required" dc:"远端服务器端口"`
	RemoteUser  string `v:"required" dc:"远端服务器用户"`
	ExtraParams string `dc:"额外传递给rsync的参数，会直接添加到rsync命令行"`
}
type SyncOutput struct{}

func (c *AutoSync) Sync(ctx context.Context, in SyncInput) (out *SyncOutput, err error) {
	// 第一次执行，强制同步
	c.doRsync(ctx, in, true)
	// 后续自动检测同步
	gtimer.AddSingleton(ctx, time.Millisecond*500, func(ctx context.Context) {
		c.doRsync(ctx, in, false)
	})
	// 启动文件监听
	c.doWatch(ctx, in)
	return
}

func (c *AutoSync) doRsync(ctx context.Context, in SyncInput, forceSync bool) {
	if !forceSync {
		content := c.writer.String()
		if content == "" {
			// 没有事件产生
			return
		}
		defer c.writer.Reset()
		fmt.Println(content)
	}

	cmd := fmt.Sprintf(
		`rsync --delete -avz -e "ssh -p %d" %s %s@%s:%s %s`,
		in.RemotePort, in.LocalDir, in.RemoteUser, in.RemoteHost, in.RemoteDir, in.ExtraParams,
	)

	g.Log().Info(ctx, cmd)
	_ = gproc.ShellRun(ctx, cmd)
	g.Log().Info(ctx, "done!\n")
}

func (c *AutoSync) doWatch(ctx context.Context, in SyncInput) {
	g.Log().Infof(ctx, `watch starts, local directory: %s`, in.LocalDir)
	cmd := gstr.Trim(fmt.Sprintf(
		`
fswatch -0 %s | while read -d "" event; do
    echo "fswatch: file ${event} has changed."
done
`,
		in.LocalDir,
	))
	p := gproc.NewProcessCmd(cmd)
	p.Stdout = c.writer
	if err := p.Run(ctx); err != nil {
		g.Log().Error(ctx, "%+v", err)
	}
}
