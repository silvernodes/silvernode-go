package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
)

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func CRC32(str string) string {
	u32 := crc32.ChecksumIEEE([]byte(str))
	return fmt.Sprint(u32)
}

func Guid() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return MD5(base64.URLEncoding.EncodeToString(b))
}
