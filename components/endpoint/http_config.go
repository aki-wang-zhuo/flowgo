package endpoint

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"strings"
)

// RouterDef 单条 HTTP 路由（方法 + 路径），对齐 RuleGo routers 的精简形态。
type RouterDef struct {
	// Name 展示名（选填）；连线标签优先用它，空则用 Path。
	Name   string `json:"name,omitempty"`
	Method string `json:"method"`
	Path   string `json:"path"`
	// DebugValue 调试用请求体 JSON（仅编辑器调试，真实 HTTP 请求不使用）。
	DebugValue string `json:"debugValue,omitempty"`
}

// HttpConfig HTTP 入口节点配置（对齐 RuleGo endpoint/http 常用字段）。
type HttpConfig struct {
	// Server 监听地址，如 ":8088" 或 "0.0.0.0:8088"。
	Server string `json:"server"`
	// AllowCors 是否允许跨域（*）。
	AllowCors bool `json:"allowCors"`
	// Https 为 true 时以 TLS 监听，须同时提供 CertPem / KeyPem。
	Https bool `json:"https"`
	// CertPem TLS 证书 PEM 文本（含 -----BEGIN CERTIFICATE-----）。
	CertPem string `json:"certPem,omitempty"`
	// KeyPem TLS 私钥 PEM 文本。
	KeyPem string `json:"keyPem,omitempty"`
	// Routers 请求路径列表。
	Routers []RouterDef `json:"routers"`
}

// DefaultHttpConfig 编辑器拖拽默认值。
func DefaultHttpConfig() HttpConfig {
	return HttpConfig{
		Server:    ":8088",
		AllowCors: true,
		Https:     false,
		Routers: []RouterDef{
			{Method: "POST", Path: "/api/demo"},
		},
	}
}

// ParseHttpConfig 从节点 configuration map 解析。
func ParseHttpConfig(cfg map[string]interface{}) (HttpConfig, error) {
	out := DefaultHttpConfig()
	if cfg == nil {
		return out, nil
	}
	if v, ok := cfg["server"].(string); ok && strings.TrimSpace(v) != "" {
		out.Server = strings.TrimSpace(v)
	}
	if v, ok := cfg["allowCors"].(bool); ok {
		out.AllowCors = v
	}
	if v, ok := cfg["https"].(bool); ok {
		out.Https = v
	}
	if v, ok := cfg["certPem"].(string); ok {
		out.CertPem = strings.TrimSpace(v)
	}
	if v, ok := cfg["keyPem"].(string); ok {
		out.KeyPem = strings.TrimSpace(v)
	}
	raw, ok := cfg["routers"]
	if !ok || raw == nil {
		if err := out.ValidateTLS(); err != nil {
			return out, err
		}
		return out, nil
	}
	list, err := parseRouters(raw)
	if err != nil {
		return out, err
	}
	if len(list) > 0 {
		out.Routers = list
	}
	if err := out.ValidateTLS(); err != nil {
		return out, err
	}
	return out, nil
}

// ValidateTLS 开启 HTTPS 时校验证书与私钥可解析。
func (c HttpConfig) ValidateTLS() error {
	if !c.Https {
		return nil
	}
	if strings.TrimSpace(c.CertPem) == "" {
		return fmt.Errorf("https 已开启，请填写证书 PEM（certPem）")
	}
	if strings.TrimSpace(c.KeyPem) == "" {
		return fmt.Errorf("https 已开启，请填写私钥 PEM（keyPem）")
	}
	if _, err := tls.X509KeyPair([]byte(c.CertPem), []byte(c.KeyPem)); err != nil {
		return fmt.Errorf("证书或私钥无效: %w", err)
	}
	return nil
}

// TLSFingerprint 用于同监听地址下比对 TLS 配置是否一致。
func (c HttpConfig) TLSFingerprint() string {
	if !c.Https {
		return ""
	}
	sum := sha256.Sum256([]byte(c.CertPem + "\x00" + c.KeyPem))
	return hex.EncodeToString(sum[:])
}

func parseRouters(raw interface{}) ([]RouterDef, error) {
	arr, ok := raw.([]interface{})
	if !ok {
		// JSON 反序列化到 map 后偶发 []map
		if typed, ok2 := raw.([]RouterDef); ok2 {
			return typed, nil
		}
		if typed, ok2 := raw.([]map[string]interface{}); ok2 {
			out := make([]RouterDef, 0, len(typed))
			for _, m := range typed {
				out = append(out, routerFromMap(m))
			}
			return out, nil
		}
		return nil, fmt.Errorf("routers must be an array")
	}
	out := make([]RouterDef, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		r := routerFromMap(m)
		if r.Path == "" {
			continue
		}
		if r.Method == "" {
			r.Method = "POST"
		}
		out = append(out, r)
	}
	return out, nil
}

func routerFromMap(m map[string]interface{}) RouterDef {
	r := RouterDef{}
	if v, ok := m["name"].(string); ok {
		r.Name = strings.TrimSpace(v)
	}
	if v, ok := m["method"].(string); ok {
		r.Method = strings.ToUpper(strings.TrimSpace(v))
	}
	if v, ok := m["path"].(string); ok {
		r.Path = strings.TrimSpace(v)
	}
	if v, ok := m["debugValue"].(string); ok {
		r.DebugValue = strings.TrimSpace(v)
	}
	return r
}

// RouterRelation 出边关系名（METHOD + 路径），供引擎按路由分支。
func RouterRelation(r RouterDef) string {
	m := strings.ToUpper(strings.TrimSpace(r.Method))
	if m == "" {
		m = "POST"
	}
	p := strings.TrimSpace(r.Path)
	if p == "" {
		p = "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return m + " " + p
}

// RouterLabel 连线展示文案：名称优先，否则 METHOD + 路径。
func RouterLabel(r RouterDef) string {
	if n := strings.TrimSpace(r.Name); n != "" {
		return n
	}
	return RouterRelation(r)
}

// NormalizeServer 规范化监听地址（仅数字时补成 :port）。
func NormalizeServer(server string) string {
	s := strings.TrimSpace(server)
	if s == "" {
		return ":8088"
	}
	if strings.HasPrefix(s, ":") {
		return s
	}
	// 纯端口数字
	onlyDigit := true
	for _, c := range s {
		if c < '0' || c > '9' {
			onlyDigit = false
			break
		}
	}
	if onlyDigit {
		return ":" + s
	}
	return s
}

// ToPattern 将路径转为 Go 1.22 ServeMux 模式；支持 {id} 与 :id。
func ToPattern(method, path string) string {
	m := strings.ToUpper(strings.TrimSpace(method))
	if m == "" {
		m = "POST"
	}
	p := strings.TrimSpace(path)
	if p == "" {
		p = "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	// :param → {param}
	parts := strings.Split(p, "/")
	for i, seg := range parts {
		if strings.HasPrefix(seg, ":") && len(seg) > 1 {
			parts[i] = "{" + seg[1:] + "}"
		}
	}
	p = strings.Join(parts, "/")
	return m + " " + p
}
