package db

import "strings"

func insertSqlStr(sqlData *SqlData) string {
	sqlname := sqlData.Name
	sqlvalue := sqlData.Value
	index := strings.LastIndex(sqlname, ",")
	if index != -1 {
		sqlname = sqlname[:index]
	}

	index = strings.LastIndex(sqlvalue, ",")
	if index != -1 {
		sqlvalue = sqlvalue[:index]
	}
	return "insert into " + sqlData.Table + " (" + sqlname + ") VALUES (" + sqlvalue + ")"
}

func deleteSqlStr(sqlData *SqlData) string {
	key := sqlData.Key
	index := strings.LastIndex(key, ",")
	if index != -1 {
		key = key[:index]
	}
	key = strings.Replace(key, ",", " and ", -1)
	return "delete from " + sqlData.Table + " where " + key
}

func updateSqlStr(sqlData *SqlData) string {
	str := sqlData.NameValue
	primary := sqlData.Key
	index := strings.LastIndex(str, ",")
	if index != -1 {
		str = str[:index]
	}

	index = strings.LastIndex(primary, ",")
	if index != -1 {
		primary = primary[:index]
	}
	primary = strings.Replace(primary, ",", " and ", -1)
	return "update " + sqlData.Table + " set " + str + " where " + primary
}

func whereSqlStr(sqlData *SqlData) string {
	key := sqlData.Key
	index := strings.LastIndex(key, ",")
	if index != -1 {
		key = key[:index]
	}
	key = strings.Replace(key, ",", " and ", -1)
	return key
}

// --- struct to sql
func InsertSql(obj interface{}, params ...OpOption) string {
	op := &Op{sqlType: SQLTYPE_INSERT}
	op.applyOpts(params)
	sqlData := &SqlData{}
	getTableName(obj, sqlData)
	parseStructSql(obj, sqlData, op)
	return insertSqlStr(sqlData)
}

func DeleteSql(obj interface{}, params ...OpOption) string {
	op := &Op{sqlType: SQLTYPE_DELETE}
	op.applyOpts(params)
	sqlData := &SqlData{}
	getTableName(obj, sqlData)
	parseStructSql(obj, sqlData, op)
	return deleteSqlStr(sqlData)
}

func UpdateSql(obj interface{}, params ...OpOption) string {
	op := &Op{sqlType: SQLTYPE_UPDATE}
	op.applyOpts(params)
	sqlData := &SqlData{}
	getTableName(obj, sqlData)
	parseStructSql(obj, sqlData, op)
	return updateSqlStr(sqlData)
}

func WhereSql(obj interface{}, params ...OpOption) string {
	params = append(params, WithOutWhere())
	op := &Op{sqlType: SQLTYPE_WHERE}
	op.applyOpts(params)
	sqlData := &SqlData{}
	getTableName(obj, sqlData)
	parseStructSql(obj, sqlData, op)
	return whereSqlStr(sqlData)
}
