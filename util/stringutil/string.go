package stringutil

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// StringArrayToString 将字符串数组连接成一个字符串，使用指定的分隔符
func StringArrayToString(arr []string, separator string) string {
	return strings.Join(arr, separator)
}

// StringToStringArray 将字符串按照指定的分隔符拆分成字符串数组
func StringToStringArray(str, separator string) []string {
	return strings.Split(str, separator)
}

// CalculateMD5 计算字符串的MD5哈希值
func CalculateMD5(input string) string {
	// 创建一个MD5哈希对象
	hasher := md5.New()

	// 将字符串转换为字节数组并计算哈希值
	hasher.Write([]byte(input))
	hash := hasher.Sum(nil)

	// 将哈希值转换为十六进制字符串
	hashString := hex.EncodeToString(hash)

	return hashString
}
