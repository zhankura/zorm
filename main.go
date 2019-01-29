package zorm

import (
	"database/sql"
	"sync"
)

// SQLCommon is the minimal database connection functionality gorm requires.  Implemented by *sql.DB.

type SQLCommon interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type sqlDb interface {
	Begin() (*sql.Tx, error)
}

type sqlTx interface {
	Commit() error
	Rollback() error
}

type DB struct {
	Value        interface{}
	Error        error
	RowsAffected int64
	db           SQLCommon
	search       *search
	values       sync.Map
}

func Open(driver string, args ...interface{}) (*DB, error) {
	var source string
	var dbSQL SQLCommon
	dbSQL, err := sql.Open(driver, source)
	if err != nil {
		return nil, err
	}
	db := &DB{
		db:        dbSQL,
		callbacks: DefaultCallback,
	}
	// Send a ping to make sure the database connection is alive.
	if d, ok := dbSQL.(*sql.DB); ok {
		if err = d.Ping(); err != nil {
			d.Close()
		}
	}
	return nil, err
}
