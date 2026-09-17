package types

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// newID 生成轻量唯一 ID（时间前缀 + 随机后缀），避免引入额外依赖。
func newID() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), hex.EncodeToString(b[:]))
}

// NewID 对外暴露的 ID 生成入口。
func NewID() string {
	return newID()
}
