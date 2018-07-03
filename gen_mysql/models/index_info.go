package models

import (
	"database/sql"
)

type IndexInfo struct {
	NonUnique    int
	IndexName    string
	ColumnName   string
	Comment      string
	IndexComment string
}

func GetIndexInfos(database_name, table_name string) (index_infos []IndexInfo, err error) {
	rows, err := MainDB.Query("SELECT NON_UNIQUE,INDEX_NAME,COLUMN_NAME,`COMMENT`,INDEX_COMMENT FROM INFORMATION_SCHEMA.STATISTICS WHERE table_schema = ? AND table_name = ?;", database_name, table_name)
	defer rows.Close()
	if err != nil {
		return index_infos, err
	}
	return _IndexInfoRowsToArray(rows)
}

func _IndexInfoRowsToArray(rows *sql.Rows) (models []IndexInfo, err error) {
	for rows.Next() {
		model := IndexInfo{}
		err = rows.Scan(&model.NonUnique, &model.IndexName, &model.ColumnName, &model.Comment, &model.IndexComment)
		if err != nil {
			return models, err
		}
		models = append(models, model)
	}
	return models, err
}
