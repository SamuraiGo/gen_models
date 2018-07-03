package gen_mysql

import (
	"bufio"
	"fmt"
	"github.com/astaxie/beego"
	"log"
	"os"
	"strings"
	"github.com/ha666/gen_models/gen_mysql/models"
	"github.com/ha666/gen_models/utils"
	"os/exec"
	"runtime"
	)

var (
	p            = ""
	name         = ""
	conn_name    = ""
	package_name = ""
	conn_string  = ""
	table_infos  []models.TableInfo
	column_infos map[string][]models.ColumnInfo
	index_infos  map[string][]models.IndexInfo
	err          error
)

func Gen(_p string) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("【Gen】ex:%v\n", err)
			return
		}
	}()
	p = _p
	load_param()
}

func load_param() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("【load_param】ex:%v\n", err)
			return
		}
	}()

	var (
		address  = ""
		port     = -1
		account  = ""
		password = ""
	)

	// 验证address参数
	{
		address = beego.AppConfig.String(p + "::address")
		if len(address) <= 0 {
			log.Fatalf("【load_param】address不能为空\n")
		}
	}

	// 验证port参数
	{
		port = beego.AppConfig.DefaultInt(p+"::port", -1)
		if port <= 1024 {
			log.Fatalf("【load_param】port不能小于1024\n")
		}
	}

	// 验证name参数
	{
		name = beego.AppConfig.String(p + "::name")
		if len(name) <= 0 {
			log.Fatalf("【load_param】name不能为空\n")
		}
	}

	// 验证account参数
	{
		account = beego.AppConfig.String(p + "::account")
		if len(account) <= 0 {
			log.Fatalf("【load_param】account不能为空\n")
		}
	}

	// 验证password参数
	{
		password = beego.AppConfig.String(p + "::password")
		if len(password) <= 0 {
			log.Fatalf("【load_param】password不能为空\n")
		}
	}

	// 验证conn_name参数
	{
		conn_name = beego.AppConfig.String(p + "::conn_name")
		if len(conn_name) <= 0 {
			log.Fatalf("【load_param】conn_name不能为空\n")
		}
	}

	// 验证package_name参数
	{
		package_name = beego.AppConfig.String(p + "::package_name")
		if len(package_name) <= 0 {
			log.Fatalf("【load_param】package_name不能为空\n")
		}
	}

	conn_string = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&loc=Asia%%2fShanghai", account, password, address, port, name)

	log.Println(conn_string)
	get_database_info()
}

func get_database_info() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("【get_database_info】ex:%v\n", err)
			return
		}
	}()

	// 连接数据库
	{
		models.InitMainDB(conn_string)
	}

	// 查询所有的表
	{
		table_infos, err = models.GetTableInfos(name)
		if err != nil {
			log.Fatalf("【get_database_info】查询表出错:%s\n", err.Error())
		}
		if len(table_infos) <= 0 || len(table_infos[0].TableName) <= 0 {
			log.Fatal("【get_database_info】查询表出错:没有找到表\n")
		}
	}

	// 加载所有的列
	{
		var column_info []models.ColumnInfo
		column_infos = make(map[string][]models.ColumnInfo, len(table_infos))
		for _, table := range table_infos {
			column_info, err = models.GetColumnInfos(name, table.TableName)
			if err != nil {
				log.Fatalf("【get_database_info】查询表%s出错:%s\n", table.TableName, err.Error())
			}
			if len(column_info) <= 0 {
				log.Fatalf("【get_database_info】查询表%s出错:没有找到列\n", table.TableName)
			}
			column_infos[table.TableName] = column_info
		}
	}

	// 加载所有的索引
	{
		var index_info []models.IndexInfo
		index_infos = make(map[string][]models.IndexInfo, len(table_infos))
		for _, table := range table_infos {
			index_info, err = models.GetIndexInfos(name, table.TableName)
			if err != nil {
				log.Fatalf("【get_database_info】查询索引%s出错:%s\n", table.TableName, err.Error())
			}
			if len(index_info) <= 0 {
				log.Fatalf("【get_database_info】查询索引%s出错:没有找到\n", table.TableName)
			}
			index_infos[table.TableName] = index_info
		}
	}

	// 处理所有的表
	{
		count := len(table_infos)
		for index, table := range table_infos {
			log.Printf("【get_database_info】%d/%d，%s", index+1, count, table.TableName)
			generator_table(&table)
		}
	}

	// 执行gofmt
	{
		os_name := runtime.GOOS
		switch os_name {
		case "darwin","linux":
			{
				cmd := exec.Command("sh", "-c", "gofmt -w ./templates/*.go")
				_, err := cmd.Output()
				if err != nil {
					log.Fatalf("【get_database_info】执行gofmt结果出错:%s\n", err.Error())
				}
			}
		case "windows":
			{
				_, err := exec.Command("cmd", "/c", `gofmt -w ./templates/*.go`).Output()
				if err != nil {
					log.Fatalf("【get_database_info】执行gofmt结果出错:%s\n", err.Error())
				}
			}
		default:
			{
				log.Println("【get_database_info】未知系统")
			}
		}
	}

}

func generator_table(table *models.TableInfo) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("【generator_table】表:%s,ex:%v\n", table.TableName, err)
			return
		}
	}()

	var (
		fd                 *os.File
		w                  *bufio.Writer
		column_info        []models.ColumnInfo
		index_info         map[string][]models.IndexInfo
		ok                 bool
		is_auto_increment  = false
		is_exist_time_type = false
		struct_name        = table.TableName
		init_index         = 0
		pk_count           = 0
		pk_column          models.ColumnInfo
	)

	// 获取列
	{
		column_info, ok = column_infos[table.TableName]
		if !ok {
			log.Fatalf("【generator_table】表:%s,没有找到列信息1", table.TableName)
		}
		if len(column_info) <= 0 {
			log.Fatalf("【generator_table】表:%s,没有找到列信息2", table.TableName)
		}
	}

	// 获取索引
	{
		index_info_s, ok := index_infos[table.TableName]
		if !ok {
			log.Fatalf("【generator_table】表:%s,没有找到索引信息1", table.TableName)
		}
		if len(index_info_s) <= 0 {
			log.Fatalf("【generator_table】表:%s,没有找到索引信息2", table.TableName)
		}
		index_info = make(map[string][]models.IndexInfo)
		for _, index := range index_info_s {
			if tmp_index, ok := index_info[index.IndexName]; ok {
				tmp_index = append(tmp_index, index)
				index_info[index.IndexName] = tmp_index
			} else {
				index_info[index.IndexName] = []models.IndexInfo{index}
			}
		}
	}

	// 判断是否有主键
	{
		is_pk := false
		for _, column := range column_info {
			if strings.EqualFold(column.ColumnKey, "PRI") {
				is_pk = true
				break
			}
		}
		if !is_pk {
			log.Fatalf("【generator_table】表:%s,没有主键", table.TableName)
		}
	}

	// 初始化操作
	{
		utils.ToBigHump(&struct_name)
		for _, column := range column_info {
			tmp_type := get_field_type(column.DataType)
			if tmp_type == "time.Time" {
				is_exist_time_type = true
			}
			if len(column.Extra) > 0 && strings.Contains(column.Extra, "auto_increment") {
				is_auto_increment = true
			}
		}
		pk_count = 0
		err = utils.PathCreate("./templates")
		if err != nil {
			log.Fatalf("【generator_table】生成表:%s,创建目录出错:%s\n", table.TableName, err.Error())
		}
		fd, err = os.OpenFile("./templates/"+table.TableName+".go", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalf("【generator_table】生成表:%s,打开文件出错:%s\n", table.TableName, err.Error())
		}
		defer fd.Close()
		fd, err = os.Create("./templates/" + table.TableName + ".go")
		if err != nil {
			log.Fatalf("【generator_table】生成表:%s,打开文件出错:%s\n", table.TableName, err.Error())
		}
		defer fd.Close()
		w = bufio.NewWriter(fd)
	}

	// Package
	{
		fmt.Fprintln(w, fmt.Sprintf("package %s\n", package_name))
		fmt.Fprintln(w, "import (")
		fmt.Fprintln(w, "\t\"database/sql\"")
		fmt.Fprintln(w, "\t\"errors\"")
		fmt.Fprintln(w, "\t\"gitee.com/ha666/golibs\"")
		if is_exist_time_type {
			fmt.Fprintln(w, "\t\"time\"")
		}
		fmt.Fprintln(w, ")")
	}

	// Struct
	{
		fmt.Fprintln(w, "")
		if len(table.TableComment) > 0 {
			fmt.Fprintln(w, fmt.Sprintf("// %s", table.TableComment))
		}
		fmt.Fprintln(w, fmt.Sprintf("type %s struct {", struct_name))
		for _, column := range column_info {
			fmt.Fprint(w, fmt.Sprintf("\t%s\t%s", column.ColumnNameCase, get_field_type(column.DataType)))
			if len(column.ColumnComment) > 0 {
				fmt.Fprint(w, fmt.Sprintf("\t//%s", column.ColumnComment))
			}
			fmt.Fprintln(w, "")
		}
		fmt.Fprintln(w, "}")
	}

	// IndexList
	{
		pk_count = 0
		for key, index := range index_info {
			if strings.EqualFold(key, "PRIMARY") {
				for _, index_a := range index {
					col, err := get_column_by_name(index_a.ColumnName, column_info)
					if err != nil {
						log.Fatalf("【generator_table】生成表:%s,IndexList查询出错:%s\n", table.TableName, err.Error())
					}
					if len(col.ColumnName) <= 0 {
						log.Fatalf("【generator_table】生成表:%s,IndexList没有找到列:%s\n", table.TableName, index_a.ColumnName)
					}
					pk_count += 1
					pk_column = col
				}
			}
		}
	}

	// Exist
	{
		fmt.Fprintln(w, "")
		fmt.Fprint(w,"// 根据【")
		for index, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if index > 0 {
				fmt.Fprint(w, ",")
			}
			if len(column.ColumnComment)>0{
				fmt.Fprint(w,column.ColumnComment)
			}else{
				fmt.Fprint(w,column.ColumnName)
			}
		}
		fmt.Fprint(w,"】查询【")
		if len(table.TableComment)>0{
			fmt.Fprint(w,table.TableComment)
		}else{
			fmt.Fprint(w,table.TableName)
		}
		fmt.Fprintln(w,"】表中是否存在相关记录")
		fmt.Fprint(w, fmt.Sprintf("func Exist%s(", struct_name))
		for index, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if index > 0 {
				fmt.Fprint(w, ",")
			}
			fmt.Fprint(w, fmt.Sprintf("%s %s", column.ColumnName, get_field_type(column.DataType)))
		}
		fmt.Fprintln(w, ") (bool, error) {")
		fmt.Fprint(w, fmt.Sprintf("\trows, err := %s.Query(\"select count(0) Count from %s where ", conn_name, table.TableName))
		for index, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if index > 0 {
				fmt.Fprint(w, " and ")
			}
			fmt.Fprint(w, fmt.Sprintf("%s=?", column.ColumnName))
		}
		fmt.Fprint(w, "\",")
		for index, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if index > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, column.ColumnName)
		}
		fmt.Fprintln(w, ")")
		fmt.Fprintln(w, "\tdefer rows.Close()")
		fmt.Fprintln(w, "\tif err != nil {")
		fmt.Fprintln(w, "\t\treturn false, err")
		fmt.Fprintln(w, "\t}")
		fmt.Fprintln(w, "\tcount := 0")
		fmt.Fprintln(w, "\tif rows.Next() {")
		fmt.Fprintln(w, "\t\terr = rows.Scan(&count)")
		fmt.Fprintln(w, "\t\tif err != nil {")
		fmt.Fprintln(w, "\t\t\treturn false, err")
		fmt.Fprintln(w, "\t\t}")
		fmt.Fprintln(w, "\t\treturn count > 0, nil")
		fmt.Fprintln(w, "\t}")
		fmt.Fprintln(w, "\treturn false, nil")
		fmt.Fprintln(w, "}")
	}

	// Insert
	{
		fmt.Fprintln(w, "")
		fmt.Fprint(w,"// 插入单条记录到【")
		if len(table.TableComment)>0{
			fmt.Fprint(w,table.TableComment)
		}else{
			fmt.Fprint(w,table.TableName)
		}
		fmt.Fprintln(w,"】表中")
		fmt.Fprint(w, fmt.Sprintf("func Insert%s(%s *%s) (", struct_name, table.TableName, struct_name))
		if is_auto_increment {
			fmt.Fprint(w, "int64")
		} else {
			fmt.Fprint(w, "bool")
		}
		fmt.Fprintln(w, ", error) {")
		fmt.Fprint(w, fmt.Sprintf("\tresult, err := %s.Exec(\"insert into %s(", conn_name, table.TableName))
		init_index = 0
		for _, column := range column_info {
			if len(column.Extra) > 0 && strings.Contains(column.Extra, "auto_increment") {
				continue
			}
			if init_index > 0 {
				fmt.Fprint(w, ",")
			}
			fmt.Fprint(w, column.ColumnName)
			init_index++
		}
		fmt.Fprint(w, ") values(")
		init_index = 0
		for _, column := range column_info {
			if len(column.Extra) > 0 && strings.Contains(column.Extra, "auto_increment") {
				continue
			}
			if init_index > 0 {
				fmt.Fprint(w, ",")
			}
			fmt.Fprint(w, "?")
			init_index++
		}
		fmt.Fprint(w, ")\",")
		init_index = 0
		for _, column := range column_info {
			if len(column.Extra) > 0 && strings.Contains(column.Extra, "auto_increment") {
				continue
			}
			if init_index > 0 {
				fmt.Fprint(w, ",")
			}
			fmt.Fprint(w, fmt.Sprintf("%s.%s", table.TableName, column.ColumnNameCase))
			init_index++
		}
		fmt.Fprintln(w, ")")
		fmt.Fprintln(w, "\tif err != nil {")
		fmt.Fprint(w, "\t\treturn ")
		if is_auto_increment {
			fmt.Fprint(w, "-1")
		} else {
			fmt.Fprint(w, "false")
		}
		fmt.Fprintln(w, ", err")
		fmt.Fprintln(w, "\t}")
		if is_auto_increment {
			fmt.Fprintln(w, "\treturn result.LastInsertId()")
		} else {
			fmt.Fprintln(w, "\taffected, err := result.RowsAffected()")
			fmt.Fprintln(w, "\tif err != nil {")
			fmt.Fprintln(w, "\t\treturn false, err")
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\treturn affected > 0, nil")
		}
		fmt.Fprintln(w, "}")
	}

	// Update
	{
		fmt.Fprintln(w, "")
		fmt.Fprint(w,"// 根据【")
		for index, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if index > 0 {
				fmt.Fprint(w, ",")
			}
			if len(column.ColumnComment)>0{
				fmt.Fprint(w,column.ColumnComment)
			}else{
				fmt.Fprint(w,column.ColumnName)
			}
		}
		fmt.Fprint(w,"】修改【")
		if len(table.TableComment)>0{
			fmt.Fprint(w,table.TableComment)
		}else{
			fmt.Fprint(w,table.TableName)
		}
		fmt.Fprintln(w,"】表的单条记录")
		fmt.Fprintln(w, fmt.Sprintf("func Update%s(%s *%s) (bool, error) {", struct_name, table.TableName, struct_name))
		fmt.Fprint(w, fmt.Sprintf("\tresult, err := %s.Exec(\"update %s set ", conn_name, table.TableName))
		init_index = 0
		for _, column := range column_info {
			if strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if init_index > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, fmt.Sprintf("%s=?", column.ColumnName))
			init_index++
		}
		fmt.Fprint(w, " where ")
		init_index = 0
		for _, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if init_index > 0 {
				fmt.Fprint(w, " and ")
			}
			fmt.Fprint(w, fmt.Sprintf("%s=?", column.ColumnName))
			init_index++
		}
		fmt.Fprint(w, "\", ")
		init_index = 0
		for _, column := range column_info {
			if strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if init_index > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, fmt.Sprintf("%s.%s", table.TableName, column.ColumnNameCase))
			init_index++
		}
		for _, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			fmt.Fprint(w, ", ")
			fmt.Fprint(w, fmt.Sprintf("%s.%s", table.TableName, column.ColumnNameCase))
		}
		fmt.Fprintln(w, ")")
		fmt.Fprintln(w, "\tif err != nil {")
		fmt.Fprintln(w, "\t\treturn false, err")
		fmt.Fprintln(w, "\t}")
		fmt.Fprintln(w, "\taffected, err := result.RowsAffected()")
		fmt.Fprintln(w, "\tif err != nil {")
		fmt.Fprintln(w, "\t\treturn false, err")
		fmt.Fprintln(w, "\t}")
		fmt.Fprintln(w, "\treturn affected > 0, nil")
		fmt.Fprintln(w, "}")
	}

	// InsertUpdate
	{
		if !is_auto_increment {
			fmt.Fprintln(w, "")
			fmt.Fprint(w,"// 插入或修改【")
			if len(table.TableComment)>0{
				fmt.Fprint(w,table.TableComment)
			}else{
				fmt.Fprint(w,table.TableName)
			}
			fmt.Fprintln(w,"】表的单条记录")
			fmt.Fprintln(w, fmt.Sprintf("func InsertUpdate%s(%s *%s) (bool, error) {", struct_name, table.TableName, struct_name))
			fmt.Fprint(w, fmt.Sprintf("\tresult, err := %s.Exec(\"insert into %s(", conn_name, table.TableName))
			init_index = 0
			for _, column := range column_info {
				if len(column.Extra) > 0 && strings.Contains(column.Extra, "auto_increment") {
					continue
				}
				if init_index > 0 {
					fmt.Fprint(w, ",")
				}
				fmt.Fprint(w, column.ColumnName)
				init_index++
			}
			fmt.Fprint(w, ") values(")
			init_index = 0
			for _, column := range column_info {
				if len(column.Extra) > 0 && strings.Contains(column.Extra, "auto_increment") {
					continue
				}
				if init_index > 0 {
					fmt.Fprint(w, ",")
				}
				fmt.Fprint(w, "?")
				init_index++
			}
			fmt.Fprint(w, ") ON DUPLICATE KEY UPDATE ")
			init_index = 0
			for _, column := range column_info {
				if strings.EqualFold(column.ColumnKey, "PRI") {
					continue
				}
				if init_index > 0 {
					fmt.Fprint(w, ",")
				}
				fmt.Fprint(w, fmt.Sprintf("%s=?", column.ColumnName))
				init_index++
			}
			fmt.Fprint(w, "\",")
			init_index = 0
			for i := 0; i < 2; i++ {
				for _, column := range column_info {
					if i > 0 && strings.EqualFold(column.ColumnKey, "PRI") {
						continue
					}
					if len(column.Extra) > 0 && strings.Contains(column.Extra, "auto_increment") {
						continue
					}
					if init_index > 0 {
						fmt.Fprint(w, ",")
					}
					fmt.Fprint(w, fmt.Sprintf("%s.%s", table.TableName, column.ColumnNameCase))
					init_index++
				}
			}
			fmt.Fprintln(w, ")")
			fmt.Fprintln(w, "\tif err != nil {")
			fmt.Fprintln(w, "\t\treturn false, err")
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\taffected, err := result.RowsAffected()")
			fmt.Fprintln(w, "\tif err != nil {")
			fmt.Fprintln(w, "\t\treturn false, err")
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\treturn affected > 0, nil")
			fmt.Fprintln(w, "}")
		}
	}

	// Delete
	{
		fmt.Fprintln(w, "")
		fmt.Fprint(w,"// 根据【")
		for index, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if index > 0 {
				fmt.Fprint(w, ",")
			}
			if len(column.ColumnComment)>0{
				fmt.Fprint(w,column.ColumnComment)
			}else{
				fmt.Fprint(w,column.ColumnName)
			}
		}
		fmt.Fprint(w,"】删除【")
		if len(table.TableComment)>0{
			fmt.Fprint(w,table.TableComment)
		}else{
			fmt.Fprint(w,table.TableName)
		}
		fmt.Fprintln(w,"】表中的单条记录")
		fmt.Fprint(w, fmt.Sprintf("func Delete%s(", struct_name))
		init_index = 0
		for _, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if init_index > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, fmt.Sprintf("%s %s", column.ColumnName, get_field_type(column.DataType)))
			init_index++
		}
		fmt.Fprintln(w, ") (bool, error) {")
		fmt.Fprint(w, fmt.Sprintf("\tresult, err := %s.Exec(\"delete from %s ", conn_name, table.TableName))
		init_index = 0
		for _, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if init_index > 0 {
				fmt.Fprint(w, " and ")
			}
			fmt.Fprint(w, fmt.Sprintf("%s=?", column.ColumnName))
			init_index++
		}
		fmt.Fprint(w, "\", ")
		init_index = 0
		for _, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if init_index > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, column.ColumnName)
			init_index++
		}
		fmt.Fprintln(w, ")")
		fmt.Fprintln(w, "\tif err != nil {")
		fmt.Fprintln(w, "\t\treturn false, err")
		fmt.Fprintln(w, "\t}")
		fmt.Fprintln(w, "\taffected, err := result.RowsAffected()")
		fmt.Fprintln(w, "\tif err != nil {")
		fmt.Fprintln(w, "\t\treturn false, err")
		fmt.Fprintln(w, "\t}")
		fmt.Fprintln(w, "\treturn affected > 0, nil")
		fmt.Fprintln(w, "}")
	}

	// DeleteIn
	{
		if pk_count == 1 {
			fmt.Fprintln(w, "")
			fmt.Fprint(w,"// 根据【")
			for index, column := range column_info {
				if !strings.EqualFold(column.ColumnKey, "PRI") {
					continue
				}
				if index > 0 {
					fmt.Fprint(w, ",")
				}
				if len(column.ColumnComment)>0{
					fmt.Fprint(w,column.ColumnComment)
				}else{
					fmt.Fprint(w,column.ColumnName)
				}
			}
			fmt.Fprint(w,"】数组删除【")
			if len(table.TableComment)>0{
				fmt.Fprint(w,table.TableComment)
			}else{
				fmt.Fprint(w,table.TableName)
			}
			fmt.Fprintln(w,"】表中的多条记录,最多20条")
			fmt.Fprintln(w, fmt.Sprintf("func Delete%sIn%s(%ss []%s) (count int64, err error) {", struct_name, pk_column.ColumnName, pk_column.ColumnName, get_field_type(pk_column.DataType)))
			fmt.Fprintln(w, fmt.Sprintf("\tif len(%ss) <= 0 {", pk_column.ColumnName))
			fmt.Fprintln(w, fmt.Sprintf("\t\treturn count, errors.New(\"%ss is empty\")", pk_column.ColumnName))
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, fmt.Sprintf("\tif len(%ss) > 20 {", pk_column.ColumnName))
			fmt.Fprintln(w, "\t\treturn count, errors.New(\"limit 20\")")
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\tsql_str := golibs.NewStringBuilder()")
			fmt.Fprintln(w, fmt.Sprintf("\tsql_str.Append(\"delete from %s\")", table.TableName))
			fmt.Fprintln(w, fmt.Sprintf("\tsql_str.Append(\" where %s in(\")", pk_column.ColumnName))
			fmt.Fprintln(w, fmt.Sprintf("\tfor i := 0; i < len(%ss); i++ {", pk_column.ColumnName))
			fmt.Fprintln(w, "\t\tif i > 0 {")
			fmt.Fprintln(w, "\t\t\tsql_str.Append(\", \")")
			fmt.Fprintln(w, "\t\t}")
			fmt.Fprintln(w, "\t\tsql_str.Append(\" ? \")")
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\tsql_str.Append(\")\")")
			fmt.Fprintln(w, "\tvar result sql.Result")
			fmt.Fprintln(w, fmt.Sprintf("\tswitch len(%ss) {", pk_column.ColumnName))
			for i := 1; i <= 20; i++ {
				fmt.Fprintln(w, fmt.Sprintf("\tcase %d:", i))
				fmt.Fprint(w, fmt.Sprintf("\t\tresult, err = %s.Exec(sql_str.ToString()", conn_name))
				for j := 0; j < i; j++ {
					fmt.Fprint(w, fmt.Sprintf(", %ss[%d]", pk_column.ColumnName, j))
				}
				fmt.Fprintln(w, ")")
			}
			fmt.Fprintln(w, "\tdefault:")
			fmt.Fprintln(w, "\t\treturn count, errors.New(\"limit 20\")")
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\tif err != nil {")
			fmt.Fprintln(w, "\t\treturn count, err")
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\taffected, err := result.RowsAffected()")
			fmt.Fprintln(w, "\tif err != nil {")
			fmt.Fprintln(w, "\t\treturn count, err")
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\treturn affected, nil")
			fmt.Fprintln(w, "}")
		}
	}

	// Get
	{
		fmt.Fprintln(w, "")
		fmt.Fprint(w,"// 根据【")
		for index, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if index > 0 {
				fmt.Fprint(w, ",")
			}
			if len(column.ColumnComment)>0{
				fmt.Fprint(w,column.ColumnComment)
			}else{
				fmt.Fprint(w,column.ColumnName)
			}
		}
		fmt.Fprint(w,"】查询【")
		if len(table.TableComment)>0{
			fmt.Fprint(w,table.TableComment)
		}else{
			fmt.Fprint(w,table.TableName)
		}
		fmt.Fprintln(w,"】表中的单条记录")
		fmt.Fprint(w, fmt.Sprintf("func Get%s(", struct_name))
		init_index = 0
		for _, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if init_index > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, fmt.Sprintf("%s %s", column.ColumnName, get_field_type(column.DataType)))
			init_index++
		}
		fmt.Fprintln(w, fmt.Sprintf(") (%s %s, err error) {", table.TableName, struct_name))
		fmt.Fprint(w, fmt.Sprintf("\trows, err := %s.Query(\"select ", conn_name))
		init_index = 0
		for _, column := range column_info {
			if init_index > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, column.ColumnName)
			init_index++
		}
		fmt.Fprint(w, fmt.Sprintf(" from %s where ", table.TableName))
		init_index = 0
		for _, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if init_index > 0 {
				fmt.Fprint(w, " and ")
			}
			fmt.Fprint(w, fmt.Sprintf("%s=?", column.ColumnName))
			init_index++
		}
		fmt.Fprint(w, "\", ")
		init_index = 0
		for _, column := range column_info {
			if !strings.EqualFold(column.ColumnKey, "PRI") {
				continue
			}
			if init_index > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, column.ColumnName)
			init_index++
		}
		fmt.Fprintln(w, ")")
		fmt.Fprintln(w, "\tdefer rows.Close()")
		fmt.Fprintln(w, "\tif err != nil {")
		fmt.Fprintln(w, fmt.Sprintf("\t\treturn %s, err", table.TableName))
		fmt.Fprintln(w, "\t}")
		fmt.Fprintln(w, fmt.Sprintf("\t%ss, err := _%sRowsToArray(rows)", table.TableName, struct_name))
		fmt.Fprintln(w, "\tif err != nil {")
		fmt.Fprintln(w, fmt.Sprintf("\t\treturn %s, err", table.TableName))
		fmt.Fprintln(w, "\t}")
		fmt.Fprintln(w, fmt.Sprintf("\tif len(%ss) <= 0 {", table.TableName))
		fmt.Fprintln(w, fmt.Sprintf("\t\treturn %s, err", table.TableName))
		fmt.Fprintln(w, "\t}")
		fmt.Fprintln(w, fmt.Sprintf("\treturn %ss[0], nil", table.TableName))
		fmt.Fprintln(w, "}")
	}

	// In
	{
		if pk_count == 1 {
			fmt.Fprintln(w, "")
			fmt.Fprint(w,"// 根据【")
			for index, column := range column_info {
				if !strings.EqualFold(column.ColumnKey, "PRI") {
					continue
				}
				if index > 0 {
					fmt.Fprint(w, ",")
				}
				if len(column.ColumnComment)>0{
					fmt.Fprint(w,column.ColumnComment)
				}else{
					fmt.Fprint(w,column.ColumnName)
				}
			}
			fmt.Fprint(w,"】数组查询【")
			if len(table.TableComment)>0{
				fmt.Fprint(w,table.TableComment)
			}else{
				fmt.Fprint(w,table.TableName)
			}
			fmt.Fprintln(w,"】表中的多条记录,最多20条")
			fmt.Fprintln(w, fmt.Sprintf("func Get%sIn%s(%ss []%s) (%ss []%s, err error) {", struct_name, pk_column.ColumnName, pk_column.ColumnName, get_field_type(pk_column.DataType), table.TableName, struct_name))
			fmt.Fprintln(w, fmt.Sprintf("\tif len(%ss) <= 0 {", pk_column.ColumnName))
			fmt.Fprintln(w, fmt.Sprintf("\t\treturn %ss, errors.New(\"%ss is empty\")", table.TableName, pk_column.ColumnName))
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, fmt.Sprintf("\tif len(%ss) > 20 {", pk_column.ColumnName))
			fmt.Fprintln(w, fmt.Sprintf("\t\treturn %ss, errors.New(\"limit 20\")", table.TableName))
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\tsql_str := golibs.NewStringBuilder()")
			fmt.Fprint(w, "\tsql_str.Append(\"select ")
			init_index = 0
			for _, column := range column_info {
				if init_index > 0 {
					fmt.Fprint(w, ", ")
				}
				fmt.Fprint(w, column.ColumnName)
				init_index++
			}
			fmt.Fprintln(w, " from \")")
			fmt.Fprintln(w, fmt.Sprintf("\tsql_str.Append(\"%s\")", table.TableName))
			fmt.Fprintln(w, fmt.Sprintf("\tsql_str.Append(\" where %s in(\")", pk_column.ColumnName))
			fmt.Fprintln(w, fmt.Sprintf("\tfor i := 0; i < len(%ss); i++ {", pk_column.ColumnName))
			fmt.Fprintln(w, "\t\tif i > 0 {")
			fmt.Fprintln(w, "\t\t\tsql_str.Append(\", \")")
			fmt.Fprintln(w, "\t\t}")
			fmt.Fprintln(w, "\t\tsql_str.Append(\" ? \")")
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\tsql_str.Append(\")\")")
			fmt.Fprintln(w, "\tvar rows *sql.Rows")
			fmt.Fprintln(w, fmt.Sprintf("\tswitch len(%ss) {", pk_column.ColumnName))
			for i := 1; i <= 20; i++ {
				fmt.Fprintln(w, fmt.Sprintf("\tcase %d:", i))
				fmt.Fprint(w, fmt.Sprintf("\t\trows, err = %s.Query(sql_str.ToString()", conn_name))
				for j := 0; j < i; j++ {
					fmt.Fprint(w, fmt.Sprintf(", %ss[%d]", pk_column.ColumnName, j))
				}
				fmt.Fprintln(w, ")")
			}
			fmt.Fprintln(w, "\tdefault:")
			fmt.Fprintln(w, fmt.Sprintf("\t\treturn %ss, errors.New(\"limit 20\")", table.TableName))
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\tdefer rows.Close()")
			fmt.Fprintln(w, "\tif err != nil {")
			fmt.Fprintln(w, fmt.Sprintf("\t\treturn %ss, err", table.TableName))
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, fmt.Sprintf("\treturn _%sRowsToArray(rows)", struct_name))
			fmt.Fprintln(w, "}")
		}
	}

	// Index
	{
		for key, index := range index_info {
			if strings.Contains(key, "PRIMARY") {
				continue
			}
			if len(index) <= 0 {
				continue
			}
			fmt.Fprintln(w, "")
			fmt.Fprint(w,"// 根据【")
			for index, index_a := range index {
				tmp_column, err := get_column_by_name(index_a.ColumnName, column_info)
				if err != nil {
					log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:%s\n", table.TableName, key, index_a.ColumnName, err.Error())
				}
				if len(tmp_column.ColumnName) <= 0 {
					log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:没有找到\n", table.TableName, key, index_a.ColumnName)
				}
				if index>0{
					fmt.Fprint(w,",")
				}
				if len(tmp_column.ColumnComment)>0{
					fmt.Fprint(w, tmp_column.ColumnComment)
				}else{
					fmt.Fprint(w, tmp_column.ColumnName)
				}
			}
			fmt.Fprint(w,"】查询【")
			if len(table.TableComment)>0{
				fmt.Fprint(w,table.TableComment)
			}else{
				fmt.Fprint(w,table.TableName)
			}
			fmt.Fprint(w,"】表中的多条记录，使用索引【")
			fmt.Fprintln(w, fmt.Sprintf("%s,%s】",index[0].IndexName,index[0].IndexComment))
			fmt.Fprint(w, fmt.Sprintf("func Get%sBy", struct_name))
			for _, index_a := range index {
				tmp_column, err := get_column_by_name(index_a.ColumnName, column_info)
				if err != nil {
					log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:%s\n", table.TableName, key, index_a.ColumnName, err.Error())
				}
				if len(tmp_column.ColumnName) <= 0 {
					log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:没有找到\n", table.TableName, key, index_a.ColumnName)
				}
				fmt.Fprint(w, tmp_column.ColumnName)
			}
			fmt.Fprint(w, "(")
			init_index = 0
			for _, index_a := range index {
				tmp_column, err := get_column_by_name(index_a.ColumnName, column_info)
				if err != nil {
					log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:%s\n", table.TableName, key, index_a.ColumnName, err.Error())
				}
				if len(tmp_column.ColumnName) <= 0 {
					log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:没有找到\n", table.TableName, key, index_a.ColumnName)
				}
				if init_index > 0 {
					fmt.Fprint(w, ", ")
				}
				fmt.Fprint(w, fmt.Sprintf("%s %s", tmp_column.ColumnName, get_field_type(tmp_column.DataType)))
				init_index++
			}
			fmt.Fprint(w, fmt.Sprintf(") (%s", table.TableName))
			if index[0].NonUnique > 0 {
				fmt.Fprint(w, "s []")
			} else {
				fmt.Fprint(w, " ")
			}
			fmt.Fprintln(w, fmt.Sprintf("%s, err error) {", struct_name))
			fmt.Fprint(w, fmt.Sprintf("\trows, err := %s.Query(\"select ", conn_name))
			init_index = 0
			for _, column := range column_info {
				if init_index > 0 {
					fmt.Fprint(w, ", ")
				}
				fmt.Fprint(w, column.ColumnName)
				init_index++
			}
			fmt.Fprint(w, fmt.Sprintf(" from %s force index(%s) where ", table.TableName, key))
			init_index = 0
			for _, index_a := range index {
				tmp_column, err := get_column_by_name(index_a.ColumnName, column_info)
				if err != nil {
					log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:%s\n", table.TableName, key, index_a.ColumnName, err.Error())
				}
				if len(tmp_column.ColumnName) <= 0 {
					log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:没有找到\n", table.TableName, key, index_a.ColumnName)
				}
				if init_index > 0 {
					fmt.Fprint(w, " and ")
				}
				fmt.Fprint(w, fmt.Sprintf("%s=?", tmp_column.ColumnName))
				init_index++
			}
			fmt.Fprint(w, "\",")
			init_index = 0
			for _, index_a := range index {
				tmp_column, err := get_column_by_name(index_a.ColumnName, column_info)
				if err != nil {
					log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:%s\n", table.TableName, key, index_a.ColumnName, err.Error())
				}
				if len(tmp_column.ColumnName) <= 0 {
					log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:没有找到\n", table.TableName, key, index_a.ColumnName)
				}
				if init_index > 0 {
					fmt.Fprint(w, ", ")
				}
				fmt.Fprint(w, tmp_column.ColumnName)
				init_index++
			}
			fmt.Fprintln(w, ")")
			fmt.Fprintln(w, "\tdefer rows.Close()")
			fmt.Fprintln(w, "\tif err != nil {")
			fmt.Fprint(w, fmt.Sprintf("\t\treturn %s", table.TableName))
			if index[0].NonUnique > 0 {
				fmt.Fprint(w, "s")
			}
			fmt.Fprintln(w, ", err")
			fmt.Fprintln(w, "\t}")
			if index[0].NonUnique == 0 {
				fmt.Fprintln(w, fmt.Sprintf("\t%ss, err := _%sRowsToArray(rows)", table.TableName, struct_name))
				fmt.Fprintln(w, "\tif err != nil {")
				fmt.Fprintln(w, fmt.Sprintf("\t\treturn %s, err", table.TableName))
				fmt.Fprintln(w, "\t}")
				fmt.Fprintln(w, fmt.Sprintf("\treturn %ss[0], nil", table.TableName))
			} else {
				fmt.Fprintln(w, fmt.Sprintf("\treturn _%sRowsToArray(rows)", struct_name))
			}
			fmt.Fprintln(w, "}")
		}

	}

	// RowCount
	{

		// 查询全部记录数
		{
			fmt.Fprintln(w, "")
			fmt.Fprint(w,"// 查询【")
			if len(table.TableComment)>0{
				fmt.Fprint(w,table.TableComment)
			}else{
				fmt.Fprint(w,table.TableName)
			}
			fmt.Fprintln(w,"】表总记录数")
			fmt.Fprintln(w, fmt.Sprintf("func Get%sRowCount() (count int, err error) {", struct_name))
			fmt.Fprintln(w, fmt.Sprintf("\trows, err := %s.Query(\"select count(0) Count from %s\")", conn_name, table.TableName))
			fmt.Fprintln(w, "\tdefer rows.Close()")
			fmt.Fprintln(w, "\tif err != nil {")
			fmt.Fprintln(w, "\t\treturn -1, err")
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\tif rows.Next() {")
			fmt.Fprintln(w, "\t\terr = rows.Scan(&count)")
			fmt.Fprintln(w, "\t\tif err != nil {")
			fmt.Fprintln(w, "\t\t\treturn -1, err")
			fmt.Fprintln(w, "\t\t}")
			fmt.Fprintln(w, "\t\treturn count, nil")
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\treturn -1, nil")
			fmt.Fprintln(w, "}")
		}

		// Index
		{
			for key, index := range index_info {
				if strings.Contains(key, "PRIMARY") {
					continue
				}
				if len(index) <= 0 {
					continue
				}
				if index[0].NonUnique == 0 {
					continue
				}
				fmt.Fprintln(w, "")
				fmt.Fprint(w,"// 根据【")
				for index, index_a := range index {
					tmp_column, err := get_column_by_name(index_a.ColumnName, column_info)
					if err != nil {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:%s\n", table.TableName, key, index_a.ColumnName, err.Error())
					}
					if len(tmp_column.ColumnName) <= 0 {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:没有找到\n", table.TableName, key, index_a.ColumnName)
					}
					if index>0{
						fmt.Fprint(w,",")
					}
					if len(tmp_column.ColumnComment)>0{
						fmt.Fprint(w, tmp_column.ColumnComment)
					}else{
						fmt.Fprint(w, tmp_column.ColumnName)
					}
				}
				fmt.Fprint(w,"】查询【")
				if len(table.TableComment)>0{
					fmt.Fprint(w,table.TableComment)
				}else{
					fmt.Fprint(w,table.TableName)
				}
				fmt.Fprint(w,"】表总记录数，使用索引【")
				fmt.Fprintln(w, fmt.Sprintf("%s,%s】",index[0].IndexName,index[0].IndexComment))
				fmt.Fprint(w, fmt.Sprintf("func Get%sRowCountBy", struct_name))
				for _, index_a := range index {
					fmt.Fprint(w, index_a.ColumnName)
				}
				fmt.Fprint(w, "(")
				init_index = 0
				for _, index_a := range index {
					tmp_column, err := get_column_by_name(index_a.ColumnName, column_info)
					if err != nil {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:%s\n", table.TableName, key, index_a.ColumnName, err.Error())
					}
					if len(tmp_column.ColumnName) <= 0 {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:没有找到\n", table.TableName, key, index_a.ColumnName)
					}
					if init_index > 0 {
						fmt.Fprint(w, ", ")
					}
					fmt.Fprint(w, fmt.Sprintf("%s %s", tmp_column.ColumnName, get_field_type(tmp_column.DataType)))
					init_index++
				}
				fmt.Fprintln(w, ") (count int, err error) {")
				fmt.Fprint(w, fmt.Sprintf("\trows, err := %s.Query(\"select count(0) Count from %s force index(%s) where ", conn_name, table.TableName, index[0].IndexName))
				init_index = 0
				for _, index_a := range index {
					tmp_column, err := get_column_by_name(index_a.ColumnName, column_info)
					if err != nil {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:%s\n", table.TableName, key, index_a.ColumnName, err.Error())
					}
					if len(tmp_column.ColumnName) <= 0 {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:没有找到\n", table.TableName, key, index_a.ColumnName)
					}
					if init_index > 0 {
						fmt.Fprint(w, " and ")
					}
					fmt.Fprint(w, fmt.Sprintf("%s=?", tmp_column.ColumnName))
					init_index++
				}
				fmt.Fprint(w, "\",")
				init_index = 0
				for _, index_a := range index {
					tmp_column, err := get_column_by_name(index_a.ColumnName, column_info)
					if err != nil {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:%s\n", table.TableName, key, index_a.ColumnName, err.Error())
					}
					if len(tmp_column.ColumnName) <= 0 {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:没有找到\n", table.TableName, key, index_a.ColumnName)
					}
					if init_index > 0 {
						fmt.Fprint(w, ", ")
					}
					fmt.Fprint(w, tmp_column.ColumnName)
					init_index++
				}
				fmt.Fprintln(w, ")")
				fmt.Fprintln(w, "\tdefer rows.Close()")
				fmt.Fprintln(w, "\tif err != nil {")
				fmt.Fprintln(w, "\t\treturn -1, err")
				fmt.Fprintln(w, "\t}")
				fmt.Fprintln(w, "\tif rows.Next() {")
				fmt.Fprintln(w, "\t\terr = rows.Scan(&count)")
				fmt.Fprintln(w, "\t\tif err != nil {")
				fmt.Fprintln(w, "\t\t\treturn -1, err")
				fmt.Fprintln(w, "\t\t}")
				fmt.Fprintln(w, "\t\treturn count, nil")
				fmt.Fprintln(w, "\t}")
				fmt.Fprintln(w, "\treturn -1, nil")
				fmt.Fprintln(w, "}")
			}
		}

	}

	// All
	{
		fmt.Fprintln(w, "")
		fmt.Fprint(w,"// 查询【")
		if len(table.TableComment)>0{
			fmt.Fprint(w,table.TableComment)
		}else{
			fmt.Fprint(w,table.TableName)
		}
		fmt.Fprintln(w,"】表所有记录")
		fmt.Fprint(w, fmt.Sprintf("func Get%sAll()", struct_name))
		fmt.Fprintln(w, fmt.Sprintf("(%ss []%s, err error) {", table.TableName, struct_name))
		fmt.Fprint(w, fmt.Sprintf("\trows, err := %s.Query(\"select ", conn_name))
		init_index = 0
		for _, column := range column_info {
			if init_index > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, column.ColumnName)
			init_index++
		}
		fmt.Fprint(w, fmt.Sprintf(" from %s", table.TableName))
		fmt.Fprintln(w, "\")")
		fmt.Fprintln(w, "\tdefer rows.Close()")
		fmt.Fprintln(w, "\tif err != nil {")
		fmt.Fprintln(w, fmt.Sprintf("\t\treturn %ss, err", table.TableName))
		fmt.Fprintln(w, "\t}")
		fmt.Fprintln(w, fmt.Sprintf("\treturn _%sRowsToArray(rows)", struct_name))
		fmt.Fprintln(w, "}")
	}

	// RowList
	{

		// NonIndex
		{
			fmt.Fprintln(w, "")
			fmt.Fprint(w,"// 分页查询【")
			if len(table.TableComment)>0{
				fmt.Fprint(w,table.TableComment)
			}else{
				fmt.Fprint(w,table.TableName)
			}
			fmt.Fprintln(w,"】表的记录")
			fmt.Fprint(w, fmt.Sprintf("func Get%sRowList", struct_name))
			fmt.Fprint(w, "(PageIndex, PageSize int) (")
			fmt.Fprintln(w, fmt.Sprintf("%ss []%s, err error) {", table.TableName, struct_name))
			fmt.Fprint(w, fmt.Sprintf("\trows, err := %s.Query(\"select ", conn_name))
			init_index = 0
			for _, column := range column_info {
				if init_index > 0 {
					fmt.Fprint(w, ", ")
				}
				fmt.Fprint(w, column.ColumnName)
				init_index++
			}
			fmt.Fprint(w, fmt.Sprintf(" from %s", table.TableName))
			fmt.Fprint(w, " limit ?,?\"")
			fmt.Fprint(w, ", (PageIndex-1)*PageSize, PageSize")
			fmt.Fprintln(w, ")")
			fmt.Fprintln(w, "\tdefer rows.Close()")
			fmt.Fprintln(w, "\tif err != nil {")
			fmt.Fprintln(w, fmt.Sprintf("\t\treturn %ss, err", table.TableName))
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, fmt.Sprintf("\treturn _%sRowsToArray(rows)", struct_name))
			fmt.Fprintln(w, "}")
		}

		// Index
		{
			for key, index := range index_info {
				if strings.Contains(key, "PRIMARY") {
					continue
				}
				if len(index) <= 0 {
					continue
				}
				if index[0].NonUnique == 0 {
					continue
				}
				fmt.Fprintln(w, "")
				fmt.Fprint(w,"// 根据【")
				for index, index_a := range index {
					tmp_column, err := get_column_by_name(index_a.ColumnName, column_info)
					if err != nil {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:%s\n", table.TableName, key, index_a.ColumnName, err.Error())
					}
					if len(tmp_column.ColumnName) <= 0 {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:没有找到\n", table.TableName, key, index_a.ColumnName)
					}
					if index>0{
						fmt.Fprint(w,",")
					}
					if len(tmp_column.ColumnComment)>0{
						fmt.Fprint(w, tmp_column.ColumnComment)
					}else{
						fmt.Fprint(w, tmp_column.ColumnName)
					}
				}
				fmt.Fprint(w,"】分页查询【")
				if len(table.TableComment)>0{
					fmt.Fprint(w,table.TableComment)
				}else{
					fmt.Fprint(w,table.TableName)
				}
				fmt.Fprint(w,"】表的记录，使用索引【")
				fmt.Fprintln(w, fmt.Sprintf("%s,%s】",index[0].IndexName,index[0].IndexComment))
				fmt.Fprint(w, fmt.Sprintf("func Get%sRowListBy", struct_name))
				for _, index_a := range index {
					fmt.Fprint(w, index_a.ColumnName)
				}
				fmt.Fprint(w, "(")
				init_index = 0
				for _, index_a := range index {
					tmp_column, err := get_column_by_name(index_a.ColumnName, column_info)
					if err != nil {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:%s\n", table.TableName, key, index_a.ColumnName, err.Error())
					}
					if len(tmp_column.ColumnName) <= 0 {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:没有找到\n", table.TableName, key, index_a.ColumnName)
					}
					if init_index > 0 {
						fmt.Fprint(w, ", ")
					}
					fmt.Fprint(w, fmt.Sprintf("%s %s", tmp_column.ColumnName, get_field_type(tmp_column.DataType)))
					init_index++
				}
				fmt.Fprint(w, ", PageIndex, PageSize int) (")
				fmt.Fprintln(w, fmt.Sprintf("%ss []%s, err error) {", table.TableName, struct_name))
				fmt.Fprint(w, fmt.Sprintf("\trows, err := %s.Query(\"select ", conn_name))
				init_index = 0
				for _, column := range column_info {
					if init_index > 0 {
						fmt.Fprint(w, ", ")
					}
					fmt.Fprint(w, column.ColumnName)
					init_index++
				}
				fmt.Fprint(w, fmt.Sprintf(" from %s force index(%s) where ", table.TableName, index[0].IndexName))
				init_index = 0
				for _, index_a := range index {
					tmp_column, err := get_column_by_name(index_a.ColumnName, column_info)
					if err != nil {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:%s\n", table.TableName, key, index_a.ColumnName, err.Error())
					}
					if len(tmp_column.ColumnName) <= 0 {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:没有找到\n", table.TableName, key, index_a.ColumnName)
					}
					if init_index > 0 {
						fmt.Fprint(w, " and ")
					}
					fmt.Fprint(w, fmt.Sprintf("%s=?", tmp_column.ColumnName))
					init_index++
				}
				fmt.Fprint(w, " limit ?,?\",")
				init_index = 0
				for _, index_a := range index {
					tmp_column, err := get_column_by_name(index_a.ColumnName, column_info)
					if err != nil {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:%s\n", table.TableName, key, index_a.ColumnName, err.Error())
					}
					if len(tmp_column.ColumnName) <= 0 {
						log.Fatalf("【generator_table】生成表:%s,index:%s,column_name:%s,err:没有找到\n", table.TableName, key, index_a.ColumnName)
					}
					if init_index > 0 {
						fmt.Fprint(w, ", ")
					}
					fmt.Fprint(w, tmp_column.ColumnName)
					init_index++
				}
				fmt.Fprint(w, ", (PageIndex-1)*PageSize, PageSize")
				fmt.Fprintln(w, ")")
				fmt.Fprintln(w, "\tdefer rows.Close()")
				fmt.Fprintln(w, "\tif err != nil {")
				fmt.Fprintln(w, fmt.Sprintf("\t\treturn %ss, err", table.TableName))
				fmt.Fprintln(w, "\t}")
				fmt.Fprintln(w, fmt.Sprintf("\treturn _%sRowsToArray(rows)", struct_name))
				fmt.Fprintln(w, "}")
			}
		}

	}

	// RowsToStruct
	{
		fmt.Fprintln(w, "")
		fmt.Fprint(w,"// 解析【")
		if len(table.TableComment)>0{
			fmt.Fprint(w,table.TableComment)
		}else{
			fmt.Fprint(w,table.TableName)
		}
		fmt.Fprintln(w,"】表记录")
		fmt.Fprintln(w, fmt.Sprintf("func _%sRowsToArray(rows *sql.Rows) (models []%s, err error) {", struct_name, struct_name))
		fmt.Fprintln(w, "\tfor rows.Next() {")
		fmt.Fprintln(w, fmt.Sprintf("\t\tmodel := %s{}", struct_name))
		fmt.Fprint(w, "\t\terr = rows.Scan(")
		init_index = 0
		for _, column := range column_info {
			if init_index > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, fmt.Sprintf("&model.%s", column.ColumnNameCase))
			init_index++
		}
		fmt.Fprintln(w, ")")
		fmt.Fprintln(w, "\t\tif err != nil {")
		fmt.Fprintln(w, "\t\t\treturn models, err")
		fmt.Fprintln(w, "\t\t}")
		fmt.Fprintln(w, "\t\tmodels = append(models, model)")
		fmt.Fprintln(w, "\t}")
		fmt.Fprintln(w, "\treturn models, err")
		fmt.Fprintln(w, "}")
	}

	// 写入文件
	{
		err = w.Flush()
		if err != nil {
			log.Fatalf("【generator_table】生成表:%s,刷新文件出错:%s\n", table.TableName, err.Error())
		}
	}

	log.Printf("【generator_table】表:%s,成功!", table.TableName)

}

func get_column_by_name(column_name string, column_info []models.ColumnInfo) (column models.ColumnInfo, err error) {
	for _, col := range column_info {
		if strings.EqualFold(col.ColumnName, column_name) {
			return col, nil
		}
	}
	return
}

func get_field_type(field_type string) string {
	type_result := ""
	switch field_type {
	case "bit":
		type_result = "bool"
	case "tinyint":
		type_result = "int8"
	case "smallint":
		type_result = "int16"
	case "int":
		type_result = "int"
	case "bigint":
		type_result = "int64"
	case "float", "decimal", "double", "numeric":
		type_result = "float64"
	case "char", "nchar", "varchar", "nvarchar", "text", "longtext", "mediumtext":
		type_result = "string"
	case "blob", "longblob", "mediumblob", "tinyblob":
		type_result = "[]byte"
	case "date", "datetime", "datetime2":
		type_result = "time.Time"
	default:
		type_result = "不支持的类型"
	}
	return type_result
}
