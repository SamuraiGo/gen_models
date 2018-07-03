package utils

import (
	"gitee.com/ha666/golibs"
	"os"
	"strings"
)

// 把下划线命名法和小驼峰命名法转成大驼峰命名法
func ToBigHump(str *string) {
	data := golibs.NewStringBuilder()
	is_capitalize := false
	for i, s := range *str {
		if string(s) == "_" {
			is_capitalize = true
		} else {
			if is_capitalize || i == 0 {
				data.Append(strings.ToUpper(string(s)))
			} else {
				data.Append(string(s))
			}
			is_capitalize = false
		}
	}
	*str = data.ToString()
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
