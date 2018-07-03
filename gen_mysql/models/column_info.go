package models

import (
	"database/sql"
	"github.com/ha666/gen_models/utils"
)

type ColumnInfo struct {
	ColumnName      string
	ColumnNameCase  string
	OrdinalPosition int
	IsNullAble      string
	DataType        string
	ColumnType      string
	ColumnKey       string
	Extra           string
	ColumnComment   string
}

func GetColumnInfos(database_name, table_name string) (column_infos []ColumnInfo, err error) {
	rows, err := MainDB.Query("SELECT COLUMN_NAME,ORDINAL_POSITION,IS_NULLABLE,DATA_TYPE,COLUMN_TYPE,COLUMN_KEY,EXTRA,COLUMN_COMMENT FROM information_schema.`COLUMNS` WHERE TABLE_SCHEMA=? AND TABLE_NAME=?;", database_name, table_name)
	defer rows.Close()
	if err != nil {
		return column_infos, err
	}
	return _ColumnInfoRowsToArray(rows)
}

func _ColumnInfoRowsToArray(rows *sql.Rows) (models []ColumnInfo, err error) {
	for rows.Next() {
		model := ColumnInfo{}
		err = rows.Scan(&model.ColumnName, &model.OrdinalPosition, &model.IsNullAble, &model.DataType, &model.ColumnType, &model.ColumnKey, &model.Extra, &model.ColumnComment)
		if err != nil {
			return models, err
		}
		model.ColumnNameCase = model.ColumnName
		utils.ToBigHump(&model.ColumnNameCase)
		models = append(models, model)
	}
	return models, err
}
