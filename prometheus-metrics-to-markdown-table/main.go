package main

import (
    "context"
    "fmt"
    "os"

    "github.com/gogf/gf/v2/os/gfile"
    "github.com/gogf/gf/v2/os/glog"
    "github.com/gogf/gf/v2/text/gregex"
)

const (
    mdTableHeader = `
| 名称 | 类型 | 描述 |
| ------ | ------ | ------ |
`
)

func main() {
    var ctx = context.Background()
    if len(os.Args) < 2 {
        glog.Fatal(ctx, `missing file path, usage eg: prometheus_metrics_to_markdown_table file.txt`)
    }
    var (
        mdContent   string
        fileContent = gfile.GetContents(os.Args[1])
    )
    // 生成markdown table
    matches, err := gregex.MatchAllString(
        `# HELP ([\s\S]+?)\s# TYPE (\w+) (\w+)`, fileContent,
    )
    if err != nil {
        panic(err)
    }
    mdContent += mdTableHeader
    for _, match := range matches {
        mdContent += fmt.Sprintf("| `%s` | `%s` | %s |\n", match[2], match[3], match[1])
    }
    fmt.Println(mdContent)
}
