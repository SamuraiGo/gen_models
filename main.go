package main

import (
	"flag"
	"fmt"
	"github.com/Unknwon/goconfig"
	"github.com/ha666/gen_models/app"
	"github.com/ha666/gen_models/gen_mysql"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

const version = "2018.0706.1530.0"

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
			if !strings.Contains(string(content), "go version") {
				log.Fatal("【main】未安装go开发环境\n")
			}
		}
	case "windows":
		{
			content, err := exec.Command("cmd", "/c", `go version`).Output()
			if err != nil {
				log.Fatalf("【main】执行go version结果出错:%s\n", err.Error())
			}
			if !strings.Contains(string(content), "go version") {
				log.Fatal("【main】未安装go开发环境\n")
			}
		}
	default:
		{
			log.Fatal("【main】未知系统")
		}
	}
	//endregion

	//region 解析配置文件
	var err error
	app.Cfg, err = goconfig.LoadConfigFile("config.ini")
	if err != nil {
		log.Fatal("【main】读取配置文件失败[config.ini]")
	}
	sections := app.Cfg.GetSectionList()
	if len(sections) <= 0 {
		log.Fatal("【main】配置文件[config.ini]配置错误")
	}
	//endregion

	//region 获取参数
	var p = flag.String("p", "", "数据库信息")
	flag.Parse()
	if len(*p) <= 0 {
		if len(sections) == 1 {
			//region 生成mysql数据库的代码
			gen_mysql.Gen(sections[0])
			//endregion
		} else {
			fmt.Printf("请选择数据库结点:(0~%d)\n", len(sections)-1)
			for index, section := range sections {
				fmt.Printf("%d、%s\n", index, section)
			}
			fmt.Print("请选择：")
			var str int
			fmt.Scanln(&str)
			if str < 0 {
				log.Fatal("【main】输入错误")
			}
			for index, section := range sections {
				if index == str {
					//region 生成mysql数据库的代码
					gen_mysql.Gen(section)
					//endregion
				}
			}
			log.Fatalf("【main】没有找到节点【%s】", str)
		}
	} else {
		for _, section := range sections {
			if section == *p {
				//region 生成mysql数据库的代码
				gen_mysql.Gen(*p)
				//endregion
			}
		}
		log.Fatalf("【main】配置文件[config.ini]中没有找到节点:%s", *p)
	}
	//endregion

	log.Println("顺利完成")
}
