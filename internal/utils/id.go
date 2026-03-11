package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/oklog/ulid/v2"
)

// NewHashID 用于生成基于内容的确定性 ID（32位 Hex）
func NewHashID(keys ...string) string {
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
	}
	hash := md5.Sum([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}

// NewRandomID 用于生成有序的全局唯一随机 ID（ULID 格式）
func NewRandomID() string {
	return ulid.Make().String()
}
