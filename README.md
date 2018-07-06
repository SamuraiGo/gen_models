# gen_models
用于生成golang中操作数据库的struct代码。

#### 使用方法：
* 在程序目录下面增加配置文件config.ini
* config.ini文件格式：
```ini
[ha666db]
address = 127.0.0.1
port = 3306
name = ha666db
account = root
password = 1234567890
conn_name = ha666db
package_name = models
```

#### 执行命令：
- windows：gen_models.exe -p ha666db (ha666db是配置文件中的section名称)
- darwin：./gen_models_darwin -p ha666db (ha666db是配置文件中的section名称)
- linux：./gen_models_linux -p ha666db (ha666db是配置文件中的section名称)
- 另外可以直接不用输入参数，windows上直接双击gen_models.exe，darwin/linux里执行./gen_models_darwin或./gen_models_linux。如果只有一个节点就直接生成代码了，多个节点会提示选择节点编号进行生成。

#### 数据库表结构建议：
- 支持：mysql数据库
- 表名：**下划线命名法**、驼峰命名法
- 字段名：**大驼峰命名法**、驼峰命名法、下划线命名法
- 所有字段不允许为空
- 一定要有主键
- 主键推荐int、bigint、varchar，这三种类型
- 索引提前创建好
