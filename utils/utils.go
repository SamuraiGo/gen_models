package utils

import (
	"os"
	"strings"
	"bytes"
)

// 把下划线命名法和小驼峰命名法转成大驼峰命名法
func ToBigHump(str *string) {
	var data bytes.Buffer
	is_capitalize := false
	for i, s := range *str {
		if string(s) == "_" {
			is_capitalize = true
		} else {
			if is_capitalize || i == 0 {
				data.WriteString(strings.ToUpper(string(s)))
			} else {
				data.WriteString(string(s))
			}
			is_capitalize = false
		}
	}
	*str = data.String()
}

// 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 创建文件夹
func PathCreate(dir string) error {
	exist, err := PathExists(dir)
	if err != nil {
		return err
	}
	if exist {
		return nil
	} else {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}
