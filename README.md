# gen_models

#### 使用方法：
* 在程序目录下面增加文件conf/app.conf
* conf/app.conf文件格式：
```ini
[ha666db]
address = "127.0.0.1"
port = 3306
name = "ha666db"
account = "root"
password = "1234567890"
conn_name = "ha666db"
package_name = "models"
```

#### 执行命令：
- windows：gen_models.exe -p ha666db (ha666db是配置文件中的section名称)
- darwin：./gen_models_darwin -p ha666db (ha666db是配置文件中的section名称)
- linux：./gen_models_linux -p ha666db (ha666db是配置文件中的section名称)

#### 数据库表结构建议：
- 支持：mysql数据库
- 表名：**下划线命名法**、驼峰命名法
- 字段名：**大驼峰命名法**、驼峰命名法、下划线命名法
- 所有字段不允许为空
- 一定要有主键
- 主键推荐int、bigint、varchar，这三种类型
- 索引提前创建好
