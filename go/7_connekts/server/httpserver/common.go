package httpserver

import (
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
	"time"
)

const SesDur = 7 * 24 * 3600

//过期时间:用户名([]uint64,如果用户名超过8字节,则会用冒号分隔各个uint64):用户名%8
func marshalCookieValue(name string) string {
	now := time.Now()
	u64 := uint64(now.Add(time.Duration(7 * 24 * 3600)*time.Second).Unix()) - uint64(now.Second()+now.Minute()*60)
	n := strconv.FormatUint(u64*13, 16)

	nameBytes := []byte(name)
	trimRight := 8-len(nameBytes) % 8
	for i := 0; i < trimRight; i++ {
		nameBytes = append(nameBytes, 0)
	}

	for i := 0; i < len(nameBytes); i += 8 {
		u64str := strconv.FormatUint(binary.LittleEndian.Uint64(nameBytes[i:i+8]), 16)
		n = n + ":" + u64str
	}
	return n +":"+strconv.Itoa(trimRight)
}

func unmarshalCookieValue(value string) (string, int64, error) {
	strs := strings.Split(value, ":")
	if len(strs) < 3 {
		return "", 0, errors.New("cookie长度错误")
	}

	lastField := len(strs)-1
	trimRight,err := strconv.Atoi(strs[lastField])
	if err != nil {
		return "", 0, errors.New("计算trimRight数失败")
	}

	unixTime, err := strconv.ParseUint(strs[0], 16, 64)
	if err != nil {
		return "", 0, errors.New("cookie解析日期失败1")
	}

	strs = strs[1:lastField]
	buf := make([]byte, len(strs)*8)

	for i := 0; i < len(strs); i++ {
		u64, err := strconv.ParseUint(strs[i], 16, 64)
		if err != nil {
			return "", 0, errors.New("cookie解析日期失败2")
		}
		binary.LittleEndian.PutUint64(buf[i*8:i*8+8], u64)
	}


	return string(buf[:len(buf)-trimRight]), int64(unixTime / 13), nil
}
