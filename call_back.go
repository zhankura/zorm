package zorm

import (
	"database/sql"
	"errors"
	"reflect"
)

var DefaultCallback = &Callback{}

type Callback struct {
	creates    []func(scope *Scope)
	updates    []func(scope *Scope)
	deletes    []func(scope *Scope)
	queries    []func(scope *Scope)
	rowQueries []func(scope *Scope)
	processors []*Callback
}

func init() {
	DefaultCallback.queries = append(DefaultCallback.queries, queryCallBack)
}

type CallbackProcesser struct {
	name      string
	before    string
	after     string
	replace   string
	remove    string
	kind      string
	processor *func(scope *Scope)
	parent    *Callback
}

type RowQueryResult struct {
	Row *sql.Row
}

type RowsQueryResult struct {
	Rows  *sql.Rows
	Error error
}

func queryCallBack(scope *Scope) {
	var (
		isSlice, isPtr bool
		resultType     reflect.Type
		results        = scope.IndirectValue()
	)

	if kind := results.Kind(); kind == reflect.Slice {
		isSlice = true
		resultType = results.Type().Elem()
		results.Set(reflect.MakeSlice(results.Type(), 0, 0))

		if resultType.Kind() == reflect.Ptr {
			isPtr = true
			resultType = resultType.Elem()
		}
	} else if kind != reflect.Struct {
		scope.Err(errors.New("unsupported destination, should be slice or struct"))
		return
	}

	scope.prepareQuerySQL()

	if !scope.HasError() {
		scope.db.RowsAffected = 0
		if rows, err := scope.db.db.Query(scope.SQL, scope.SQLVars...); scope.Err(err) == nil {
			columns, _ := rows.Columns()
			for rows.Next() {
				scope.db.RowsAffected++

				elem := results
				if isSlice {
					elem = reflect.New(resultType).Elem()
				}

				scope.scan(rows, columns, scope.New(elem.Addr().Interface()).Fields())

				if isSlice {
					if isPtr {
						results.Set(reflect.Append(results, elem.Addr()))
					} else {
						results.Set(reflect.Append(results, elem))
					}
				}
			}

			if err := rows.Err(); err != nil {
				scope.Err(err)
			}
		}
	}

}

func insertCallBack(scope *Scope) {}

func updateCallBack(scope *Scope) {}

func deleteCallBack(scope *Scope) {}
