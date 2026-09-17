/**
 * 内置节点编辑器文档：从同目录 Markdown 文件嵌入。
 * 命名约定：{节点类型}_zh.md / {节点类型}_en.md（如 inject_zh.md）。
 * 与 ComponentDef.Usage（MCP/AI）分离；仅供属性面板「文档」页。
 * 正文内勿出现反引号；代码块用 ~~~ 围栏。
 */
package nodedocs

import (
	"embed"
	"io/fs"
	"strings"

	"github.com/flowgo/flowgo/api/types"
)

//go:embed *_zh.md *_en.md
var mdFS embed.FS

// BuiltinMap 返回 type -> locale -> Markdown 正文（编译期嵌入）。
func BuiltinMap() map[string]map[string]string {
	out := map[string]map[string]string{}
	_ = fs.WalkDir(mdFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		typeName, locale, ok := ParseDocFileName(d.Name())
		if !ok {
			return nil
		}
		b, err := mdFS.ReadFile(path)
		if err != nil {
			return nil
		}
		body := strings.TrimSpace(string(b))
		if body == "" {
			return nil
		}
		if out[typeName] == nil {
			out[typeName] = map[string]string{}
		}
		out[typeName][locale] = body
		return nil
	})
	return out
}

// ParseDocFileName 解析 {type}_zh.md / {type}_en.md。
// 返回节点类型、规范化 locale、是否匹配。
func ParseDocFileName(name string) (typeName, locale string, ok bool) {
	base := strings.TrimSpace(name)
	base = strings.ReplaceAll(base, "\\", "/")
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	lower := strings.ToLower(base)
	if !strings.HasSuffix(lower, ".md") {
		return "", "", false
	}
	stem := base[:len(base)-3] // 去掉 .md，保留原始大小写 type
	switch {
	case strings.HasSuffix(strings.ToLower(stem), "_zh"):
		typeName = stem[:len(stem)-3]
		locale = types.LocaleZhCN
	case strings.HasSuffix(strings.ToLower(stem), "_en"):
		typeName = stem[:len(stem)-3]
		locale = types.LocaleEnUS
	default:
		return "", "", false
	}
	typeName = strings.TrimSpace(typeName)
	if typeName == "" {
		return "", "", false
	}
	return typeName, locale, true
}

// IsDocMarkdown 判断文件名是否为节点文档 Markdown。
func IsDocMarkdown(name string) bool {
	_, _, ok := ParseDocFileName(name)
	return ok
}
