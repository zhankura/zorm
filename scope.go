package zorm

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Scope struct {
	Search     *search
	Value      interface{}
	SQL        string
	SQLVars    []interface{}
	db         *DB
	instanceID string
	fields     *[]*Field
}

func (scope *Scope) IndirectValue() reflect.Value {
	return indirect(reflect.ValueOf(scope.Value))
}

func (scope *Scope) HasError() bool {
	return scope.db.Error != nil
}
func (scope *Scope) Set(name string, value interface{}) *Scope {
	scope.db.InstantSet(name, value)
	return scope
}

func (scope *Scope) Get(name string) (interface{}, bool) {
	return scope.db.Get(name)
}

func (scope *Scope) Err(err error) error {
	if err != nil {
		scope.db.AddError(err)
	}
	return err
}

func (scope *Scope) New(value interface{}) *Scope {
	return &Scope{db: scope.NewDB(), Search: &search{}, Value: value}
}

func (scope *Scope) NewDB() *DB {
	if scope.db != nil {
		db := scope.db.clone()
		db.search = nil
		db.Value = nil
		return db
	}
	return nil
}

func (scope *Scope) InstanceID() string {
	if scope.instanceID == "" {
		scope.instanceID = fmt.Sprintf("%v%v", &scope, &scope.db)
	}
	return scope.instanceID
}

func (scope *Scope) InstanceSet(name string, value interface{}) *Scope {
	return scope.Set(name+scope.InstanceID(), value)
}

func (scope *Scope) InstanceGet(name string) (interface{}, bool) {
	return scope.Get(name + scope.InstanceID())
}

func (scope *Scope) Fields() []*Field {
	if scope.fields == nil {
		var (
			fields             []*Field
			indirectScopeValue = scope.IndirectValue()
			isStruct           = indirectScopeValue.Kind() == reflect.Struct
		)
		for _, structField := range scope.GetModelStruct().StructFields {
			if isStruct {
				fieldValue := indirectScopeValue
				name := structField.Name
				if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
					fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
				}
				fieldValue = reflect.Indirect(fieldValue).FieldByName(name)
				fields = append(fields, &Field{StructField: structField, Field: fieldValue})
			} else {
				fields = append(fields, &Field{StructField: structField, IsBlank: true})
			}
		}
		scope.fields = &fields
	}
	return *scope.fields
}

func (scope *Scope) FieldByName(fieldName string) (*Field, bool) {
	quoteFieldName := toColumnName(fieldName)
	for _, field := range scope.Fields() {
		if field.Name == quoteFieldName {
			return field, true
		}
	}
	return nil, false
}
func (scope *Scope) AddToVars(value interface{}) string {
	if expr, ok := value.(*expr); ok {
		exp := expr.expr
		return exp
	}
	scope.SQLVars = append(scope.SQLVars, value)
	return "$$$"
}

func (scope *Scope) QuotedTableName() (name string) {
	if scope.Search != nil && len(scope.Search.tableName) > 0 {
		return scope.Search.tableName
	}
	return scope.GetModelStruct().TableName()
}

func (scope *Scope) Raw(sql string) *Scope {
	scope.SQL = strings.Replace(sql, "$$$", "?", -1)
	return scope
}

func (scope *Scope) CombinedConditionSql() string {
	whereSQL := scope.whereSQL()
	if scope.Search.raw {
		whereSQL = strings.TrimPrefix(strings.TrimSuffix(whereSQL, ")"), "WHERE (")
	}
	return scope.joinSQL() + whereSQL + scope.groupSQL() + scope.havingSQL() + scope.orderSQL() + scope.limitAndOffsetSQL()
}

var (
	columnRegexp        = regexp.MustCompile("^[a-zA-Z\\d]+(\\.[a-zA-Z\\d]+)*$") // only match string like `name`, `users.name`
	isNumberRegexp      = regexp.MustCompile("^\\s*\\d+\\s*$")                   // match if string is number
	comparisonRegexp    = regexp.MustCompile("(?i) (=|<>|(>|<)(=?)|LIKE|IS|IN) ")
	countingQueryRegexp = regexp.MustCompile("(?i)^count(.+)$")
)

func (scope *Scope) callCallbacks(funcs []func(s *Scope)) *Scope {
	defer func() {
		if err := recover(); err != nil {
			if db, ok := scope.db.db.(sqlTx); ok {
				db.Rollback()
			}
			panic(err)
		}
	}()
	for _, f := range funcs {
		f(scope)
	}
	return scope
}

func (scope *Scope) row() *sql.Row {
	result := &RowQueryResult{}
	scope.InstanceSet("row_query_result", result)
	scope.callCallbacks(scope.db.callbacks.rowQueries)
	return result.Row
}

func (scope *Scope) rows() (*sql.Rows, error) {
	result := &RowsQueryResult{}
	scope.InstanceSet("row_query_result", result)
	scope.callCallbacks(scope.db.callbacks.rowQueries)
	return result.Rows, result.Error
}

func (scope *Scope) prepareQuerySQL() {
	if scope.Search.raw {
		scope.Raw(scope.CombinedConditionSql())
	} else {
		scope.Raw(fmt.Sprintf("SELECT %v FROM %v %v", scope.selectSQL(), scope.QuotedTableName(), scope.CombinedConditionSql()))
	}
}

func (scope *Scope) groupSQL() string {
	if len(scope.Search.group) == 0 {
		return ""
	}
	return " GROUP BY " + scope.Search.group
}

func (scope *Scope) havingSQL() string {
	if len(scope.Search.havingConditions) == 0 {
		return ""
	}
	var andConditions []string
	for _, clause := range scope.Search.havingConditions {
		if sql := scope.buildCondition(clause, true); sql != "" {
			andConditions = append(andConditions, sql)
		}
	}

	if len(andConditions) == 0 {
		return ""
	}
	combinedSQL := strings.Join(andConditions, " AND ")

	return combinedSQL
}

func (scope *Scope) whereSQL() (sql string) {
	var (
		andConditions, orConditions       []string
		deletedAtField, hasDeletedAtField = scope.FieldByName("DeletedAt")
		quoteTableName                    = scope.QuotedTableName()
	)

	for _, clause := range scope.Search.whereConditions {
		if sql := scope.buildCondition(clause, true); sql != "" {
			andConditions = append(andConditions, sql)
		}
	}

	for _, clause := range scope.Search.orConditions {
		if sql := scope.buildCondition(clause, true); sql != "" {
			orConditions = append(orConditions, sql)
		}
	}

	for _, clause := range scope.Search.notConditions {
		if sql := scope.buildCondition(clause, false); sql != "" {
			andConditions = append(andConditions, sql)
		}
	}

	if scope.Search.Unscoped && hasDeletedAtField {
		sql := fmt.Sprintf("%v.%v IS NULL", quoteTableName, deletedAtField.DBName)
		andConditions = append(andConditions, sql)
	}

	orSQL := strings.Join(orConditions, " OR ")
	combinedSQL := strings.Join(andConditions, " AND ")
	if len(combinedSQL) > 0 {
		if len(orSQL) > 0 {
			combinedSQL = combinedSQL + " OR " + orSQL
		}
	} else {
		combinedSQL = orSQL
	}

	if len(combinedSQL) > 0 {
		sql = "WHERE " + combinedSQL
	}

	return
}

func (scope *Scope) selectSQL() string {
	if scope.Search.selects == "" {
		return "*"
	}
	return scope.Search.selects
}

func (scope *Scope) orderSQL() string {
	if len(scope.Search.orders) == 0 {
		return ""
	}
	var orders []string
	for _, order := range scope.Search.orders {
		if str, ok := order.(string); ok {
			orders = append(orders, str)
		} else {
			scope.Err(fmt.Errorf("invalid query condition: %v", order))
			return ""
		}
	}
	return " ORDER BY " + strings.Join(orders, ",")
}

func (scope *Scope) joinSQL() string {
	var joinConditions []string
	for _, clause := range scope.Search.joinConditions {
		if sql := scope.buildCondition(clause, true); sql != "" {
			joinConditions = append(joinConditions, strings.TrimSuffix(strings.TrimPrefix(sql, "("), ")"))
		}
	}
	return strings.Join(joinConditions, " ") + " "
}

func (scope *Scope) limitAndOffsetSQL() (sql string) {
	if scope.Search.limit != nil {
		if parsedLimit, err := strconv.ParseInt(fmt.Sprint(scope.Search.limit), 0, 0); err == nil && parsedLimit >= 0 {
			sql += fmt.Sprintf(" LIMIT %d", parsedLimit)
		}
	}
	if scope.Search.offset != nil {
		if parsedOffset, err := strconv.ParseInt(fmt.Sprint(scope.Search.offset), 0, 0); err == nil && parsedOffset >= 0 {
			sql += fmt.Sprintf(" OFFSET %d", parsedOffset)
		}
	}
	return
}

func (scope *Scope) buildCondition(clause map[string]interface{}, include bool) (str string) {
	var (
		tableName = scope.Search.tableName
		equalSQL  = "="
	)

	if !include {
		equalSQL = "<>"
	}

	switch value := clause["query"].(type) {
	case string:
		if value != "" {
			if !include {
				if comparisonRegexp.MatchString(value) {
					str = fmt.Sprintf("NOT (%v)", value)
				} else {
					str = fmt.Sprintf("%v.%v NOT IN (?)", tableName, value)
				}
			} else {
				str = fmt.Sprintf("(%v)", value)
			}
		}
	case map[string]interface{}:
		var sqls []string
		for key, value := range value {
			if value != nil {
				sqls = append(sqls, fmt.Sprintf("(%v.%v %s %v)", tableName, key, equalSQL, scope.AddToVars(value)))
			} else {
				if !include {
					sqls = append(sqls, fmt.Sprintf("(%v.%v IS NOT NULL)", tableName, key))
				} else {
					sqls = append(sqls, fmt.Sprintf("(%v.%v IS NULL)", tableName, key))
				}
			}
		}
		return strings.Join(sqls, " AND ")
	default:
		scope.Err(fmt.Errorf("invalid query condition: %v", value))
		return
	}

	replacements := make([]string, 0)
	args := clause["args"].([]interface{})
	for _, arg := range args {
		switch reflect.ValueOf(arg).Kind() {
		case reflect.Slice:
			if values := reflect.ValueOf(arg); values.Len() > 0 {
				tempMarks := make([]string, 0)
				for i := 0; i < values.Len(); i++ {
					tempMarks = append(tempMarks, scope.AddToVars(values.Index(i).Interface()))
				}
				replacements = append(replacements, strings.Join(tempMarks, ","))
			} else {
				replacements = append(replacements, scope.AddToVars(Expr("NULL")))
			}
		default:
			replacements = append(replacements, scope.AddToVars(arg))
		}
	}

	buff := bytes.NewBuffer([]byte{})
	i := 0
	for _, s := range str {
		if s == '?' && len(replacements) > i {
			buff.WriteString(replacements[i])
			i++
		} else {
			buff.WriteRune(s)
		}
	}

	str = buff.String()

	return

}

func indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func (scope *Scope) scan(rows *sql.Rows, columns []string, fields []*Field) {
	var (
		ignored      interface{}
		values       = make([]interface{}, len(columns))
		selectFields []*Field
		resetFields  = map[int]*Field{}
	)

	for index, column := range columns {
		values[index] = &ignored
		selectFields = fields

		for _, field := range selectFields {
			if field.DBName == column {
				if field.Field.Kind() == reflect.Ptr {
					values[index] = field.Field.Addr().Interface()
				} else {
					reflectValue := reflect.New(reflect.PtrTo(field.Struct.Type))
					reflectValue.Elem().Set(field.Field.Addr())
					values[index] = reflectValue.Interface()
					resetFields[index] = field
				}
				if field.IsNormal {
					break
				}
			}
		}
	}

	scope.Err(rows.Scan(values...))

	for index, field := range resetFields {
		if v := reflect.ValueOf(values[index]).Elem().Elem(); v.IsValid() {
			field.Field.Set(v)
		}
	}

}
