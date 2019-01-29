package zorm

type Scope struct {
	Search     *search
	Value      interface{}
	SQL        string
	SQLVars    []interface{}
	db         *DB
	instanceID string
	fields     *[]*Field
}
