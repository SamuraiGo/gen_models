package main

import (
	"flag"
	"log"
	"github.com/ha666/gen_models/gen_mysql"
)

const version = "2018.0703.1800.0"

func main() {
	log.Printf("当前版本：%s\n", version)
	var p = flag.String("p", "", "数据库信息")
	flag.Parse()
	if len(*p) <= 0 {
		log.Fatal("没有找到p参数\n")
	}
	gen_mysql.Gen(*p)
	log.Println("顺利完成")
}
