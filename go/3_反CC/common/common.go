package common

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net"
)

func Md5sum(key string) string {
	w := md5.New()
	io.WriteString(w, key)
	return hex.EncodeToString(w.Sum(nil))
}

func SplitHostPort(addr string, index int) string {
	h, p, _ := net.SplitHostPort(addr)
	if index == 0 {
		return h
	}
	return p
}
