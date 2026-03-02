package md5

import (
	"crypto/md5"
	"fmt"
	"io"
)

func Md5(message string) string {
	m := md5.New()
	_, err := io.WriteString(m, message)
	if err != nil {
		return ""
	}
	arr := m.Sum(nil)
	return fmt.Sprintf("%x", arr)
}
