package utils

import (
	"regexp"
	"strings"
)

// 添加一个新的辅助函数来处理文件名
func CleanShortcutName(filename string) string {
	// 移除 .lnk 后缀
	name := strings.TrimSuffix(filename, ".lnk")

	// 使用正则表达式移除 ".exe - 快捷方式" 模式
	re := regexp.MustCompile(`(.+)\.exe.*$`)
	if match := re.FindStringSubmatch(name); len(match) > 1 {
		name = match[1]
	}

	return name
}

func IsShortcut(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".lnk")
}
