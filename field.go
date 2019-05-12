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

func (sf *StructField) TagSettingSet(key, val string) {
	sf.tagSettingsLock.Lock()
	defer sf.tagSettingsLock.Unlock()
	sf.TagSettings[key] = val
}

func (sf *StructField) TagSettingGet(key string) (string, bool) {
	sf.tagSettingsLock.Lock()
	defer sf.tagSettingsLock.Unlock()
	val, ok := sf.TagSettings[key]
	return val, ok
}
