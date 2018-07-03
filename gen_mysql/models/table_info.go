package models

import "database/sql"

type TableInfo struct {
	TableName    string
	TableType    string
	Engine       string
	TableComment string
}

func GetTableInfos(database_name string) (table_infos []TableInfo, err error) {
	rows, err := MainDB.Query("SELECT TABLE_NAME,TABLE_TYPE,`ENGINE`,TABLE_COMMENT FROM information_schema.`TABLES` WHERE TABLE_SCHEMA=?", database_name)
	defer rows.Close()
	if err != nil {
		return table_infos, err
	}
	return _TableInfoRowsToArray(rows)
}

func _TableInfoRowsToArray(rows *sql.Rows) (models []TableInfo, err error) {
	for rows.Next() {
		model := TableInfo{}
		err = rows.Scan(&model.TableName, &model.TableType, &model.Engine, &model.TableComment)
		if err != nil {
			return models, err
		}
		models = append(models, model)
	}
	return models, err
}
