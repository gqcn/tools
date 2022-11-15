package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gproc"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/gogf/gf/v2/text/gstr"
)

const (
	localDir   = `/Users/john/Workspace/Go/GOPATH/src/git.code.oa.com/Khaos/`
	remoteDir  = `/root/workspace/khaos`
	remoteHost = `DevCloud`
	remotePort = 36000
	remoteUser = `root`
)

var (
	ctx    = gctx.GetInitCtx()
	writer = NewWriter()
)

func doRsync(ctx context.Context) {
	content := writer.String()
	if content == "" {
		return
	}
	defer writer.Reset()

	fmt.Println(content)

	cmd := fmt.Sprintf(
		`rsync --delete -avz -e "ssh -p %d" %s %s@%s:%s`,
		remotePort, localDir, remoteUser, remoteHost, remoteDir,
	)
	g.Log().Info(ctx, cmd)
	_ = gproc.ShellRun(ctx, cmd)
	g.Log().Info(ctx, "done!\n")
}

func doWatch() {
	g.Log().Infof(ctx, `watch starts, local directory: %s`, localDir)
	cmd := fmt.Sprintf(
		`
fswatch -0 %s | while read -d "" event; do
    echo "fswatch: file ${event} has changed."
done
`,
		localDir,
	)
	p := gproc.NewProcessCmd(gstr.Trim(cmd))
	p.Stdout = writer
	if err := p.Run(ctx); err != nil {
		g.Log().Error(ctx, "%+v", err)
	}
}

func main() {
	gtimer.AddSingleton(ctx, time.Second, doRsync)
	doWatch()
}
