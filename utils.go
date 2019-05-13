package zorm

import "strings"

type expr struct {
	expr string
	args []interface{}
}

func Expr(expression string, args ...interface{}) *expr {
	return &expr{expr: expression, args: args}
}

func toTableName(rawTableName string) string {
	data := make([]byte, 0, len(rawTableName)*2)
	j := false
	num := len(rawTableName)
	for i := 0; i < num; i++ {
		d := rawTableName[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	tableName := strings.ToLower(string(data[:]))
	if !strings.HasSuffix(tableName, "s") {
		tableName = tableName + "s"
	}
	return tableName
}

func toColumnName(name string) string {
	data := make([]byte, 0, len(name)*2)
	j := false
	for i := 0; i < len(name); i++ {
		d := name[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}
