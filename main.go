package main

import (
	"flag"
	"log"
	"github.com/ha666/gen_models/gen_mysql"
	"runtime"
	"os/exec"
	"strings"
)

const version = "2018.0706.0930.0"

func main() {

	//region 输出当前版本
	log.Printf("当前版本：%s\n", version)
	//endregion

	//region 判断是否安装go开发环境
	os_name := runtime.GOOS
	switch os_name {
	case "darwin", "linux":
		{
			cmd := exec.Command("sh", "-c", "go version")
			content, err := cmd.Output()
			if err != nil {
				log.Fatalf("【main】执行go version结果出错:%s\n", err.Error())
			}
			if !strings.Contains(string(content),"go version"){
				log.Fatal("【main】未安装go开发环境\n")
			}
		}
	case "windows":
		{
			content, err := exec.Command("cmd", "/c", `gofmt -w templates`).Output()
			if err != nil {
				log.Fatalf("【main】执行go version结果出错:%s\n", err.Error())
			}
			if !strings.Contains(string(content),"go version"){
				log.Fatal("【main】未安装go开发环境\n")
			}
		}
	default:
		{
			log.Fatal("【main】未知系统")
		}
	}
	//endregion

	//region 获取参数
	var p = flag.String("p", "", "数据库信息")
	flag.Parse()
	if len(*p) <= 0 {
		log.Fatal("没有找到p参数\n")
	}
	//endregion

	//region 生成mysql数据库的代码
	gen_mysql.Gen(*p)
	//endregion

	log.Println("顺利完成")
}
