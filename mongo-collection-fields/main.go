package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CollectionSpec 集合信息。
type CollectionSpec struct {
	Name    string                 // 集合名称
	Example string                 // 集合代码示例
	Comment string                 // 集合描述说明
	Fields  []*CollectionSpecField // 字段列表
}

// CollectionSpecField 字段信息。
type CollectionSpecField struct {
	Name    string // 字段名称
	Type    string // 字段类型
	Path    string // 字段层级关系，用于embedded document，例如 buildinfo.version
	Indent  string // 字段缩进，主要用于优雅展示字段层级关系，用于embedded document
	Comment string // 字段描述说明
	Example string // 字段代码示例
}

const (
	generatedPath = "/Users/john/Temp/api.MD"
	basicIndent   = " - "
)

func main() {
	var (
		ctx      = context.Background()
		mongoDbs = g.Cfg().MustGet(ctx, "mongodb.dbs").Strings()
		mongoUri = g.Cfg().MustGet(ctx, "mongodb.uri").String()
	)
	mongoClient, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(mongoUri),
	)
	if err != nil {
		g.Log().Fatal(ctx, err)
	}
	var (
		cps  = make([]*CollectionSpec, 0)
		opts = options.Find().SetSort(bson.D{{"_id", -1}}).SetLimit(1)
	)
	for _, dbName := range mongoDbs {
		collections, err := mongoClient.Database(dbName).ListCollectionNames(ctx, bson.D{})
		if err != nil {
			g.Log().Fatal(ctx, err)
		}
		for _, collection := range collections {
			g.Log().Infof(ctx, `generating collection spec for "%s"`, collection)
			cur, err := mongoClient.
				Database(dbName).
				Collection(collection).
				Find(ctx, bson.D{}, opts)
			if err != nil {
				g.Log().Warning(ctx, err)
				continue
			}
			cur.Next(ctx)

			var record = cur.Current
			var cp = &CollectionSpec{
				Name:    collection,
				Fields:  make([]*CollectionSpecField, 0),
				Example: prettyRaw(record),
			}
			listDocument(cp, "", "", record)
			cps = append(cps, cp)
		}
	}
	generateMd(cps)
}

func generateMd(cps []*CollectionSpec) {
	var (
		buffer = bytes.NewBuffer(nil)
	)
	for _, cp := range cps {
		// 忽略没有字段的集合
		if len(cp.Fields) == 0 {
			continue
		}
		buffer.WriteString(fmt.Sprintf("# %s\n", cp.Name))
		buffer.WriteString("<!-- 集合的说明描述信息写在这里 -->\n")
		buffer.WriteString("<!-- 🔥注意：除了集合和字段的说明描述信息，其他信息都会被自动化同步覆盖 -->\n")
		buffer.WriteString("\n")
		buffer.WriteString(fmt.Sprintf("```json\n%s\n```\n", cp.Example))
		buffer.WriteString("| Field | Type | Comment | Example|\n")
		buffer.WriteString("|---|---|---|---|\n")
		for _, field := range cp.Fields {
			// 忽略键名为数字的场景
			if gstr.IsNumeric(field.Name) {
				continue
			}
			buffer.WriteString(fmt.Sprintf("|%s`%s`", field.Indent, field.Name))
			buffer.WriteString(fmt.Sprintf("|`%s`", field.Type))
			buffer.WriteString(fmt.Sprintf("|%s", field.Comment))
			buffer.WriteString(fmt.Sprintf("|<pre>%s</pre>|\n", string2md(field.Example)))
		}
		buffer.WriteString("\n")
	}
	_ = gfile.PutContents(generatedPath, buffer.String())
}

// indent 用于代码缩进展示；path 表示层级关系。
func listDocument(cp *CollectionSpec, indent, path string, doc bson.Raw) {
	var (
		recordMap   = doc2Map(doc)
		elements, _ = doc.Elements()
	)
	for _, element := range elements {
		var (
			key   = element.Key()
			value = element.Value()
		)
		var (
			newIndent = indent + basicIndent
			newPath   = fmt.Sprintf(`%s.%s`, path, key)
			field     = &CollectionSpecField{
				Name:   key,
				Type:   value.Type.String(),
				Indent: indent,
				Path:   newPath,
			}
		)
		cp.Fields = append(cp.Fields, field)
		switch element.Value().Type {
		case bson.TypeEmbeddedDocument:
			var embeddedDoc = element.Value().Document()
			field.Example = prettyRaw(embeddedDoc)
			listDocument(cp, newIndent, newPath, embeddedDoc)
		default:
			field.Example = gconv.String(recordMap[element.Key()])
		}
	}
}

func doc2Map(raw bson.Raw) bson.M {
	if raw == nil {
		return bson.M{}
	}
	rawEncoded, err := bson.Marshal(raw)
	if err != nil {
		panic(err)
	}
	var m bson.M
	if err = bson.Unmarshal(rawEncoded, &m); err != nil {
		panic(err)
	}
	return m
}

func prettyRaw(raw bson.Raw) string {
	var m = doc2Map(raw)
	prettyJSON, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(prettyJSON)
}

func string2md(s string) string {
	return gstr.ReplaceByMap(s, map[string]string{
		" ":  "&nbsp;",
		"\n": "<br/>",
	})
}
