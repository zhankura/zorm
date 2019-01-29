package zorm

import (
	"reflect"
	"sync"
)

type Field struct {
	*StructField
	IsBlank bool
	Field   reflect.Value
}

type StructField struct {
	DBName          string
	Name            string
	IsPrimaryKey    bool
	IsNormal        bool
	IsScanner       bool
	HasDefaultValue bool
	Tag             reflect.StructTag
	TagSettings     map[string]string
	Struct          reflect.StructField

	tagSettingsLock sync.RWMutex
}
