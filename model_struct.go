package zorm

import (
	"go/ast"
	"reflect"
	"strings"
	"sync"
	"time"
)

type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

var modelStructsMap sync.Map

type ModelStruct struct {
	PrimaryFields []*StructField
	StructFields  []*StructField
	ModelType     reflect.Type

	defaultTableName string
	l                sync.Mutex
}

type tabler interface {
	TableName() string
}

func (s *ModelStruct) TableName() string {
	s.l.Lock()
	defer s.l.Unlock()

	if s.defaultTableName == "" && s.ModelType != nil {
		if tabler, ok := reflect.New(s.ModelType).Interface().(tabler); ok {
			s.defaultTableName = tabler.TableName()
		} else {
			tableName := toTableName(s.ModelType.Name())
			s.defaultTableName = tableName
		}
	}
	return s.defaultTableName
}

func (scope *Scope) GetModelStruct() *ModelStruct {
	var modelStruct ModelStruct
	if scope.Value == nil {
		return &modelStruct
	}
	reflectType := reflect.ValueOf(scope.Value).Type()
	for reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	if reflectType.Kind() != reflect.Struct {
		return &modelStruct
	}
	if value, ok := modelStructsMap.Load(reflectType); ok && value != nil {
		return value.(*ModelStruct)
	}
	modelStruct.ModelType = reflectType
	for i := 0; i < reflectType.NumField(); i++ {
		if fieldStruct := reflectType.Field(i); ast.IsExported(fieldStruct.Name) {
			field := &StructField{
				Struct:      fieldStruct,
				Name:        fieldStruct.Name,
				Tag:         fieldStruct.Tag,
				TagSettings: parseTagSetting(fieldStruct.Tag),
			}
			if _, ok := field.TagSettingGet("PRIMARY_KEY"); ok {
				field.IsPrimaryKey = true
				modelStruct.PrimaryFields = append(modelStruct.PrimaryFields, field)
			}
			if _, ok := field.TagSettingGet("DEFAULT"); ok {
				field.HasDefaultValue = true
			}
			if _, ok := field.TagSettingGet("AUTO_INCREMENT"); ok && !field.IsPrimaryKey {
				field.HasDefaultValue = true
			}
			if value, ok := field.TagSettingGet("COLUMN"); ok {
				field.DBName = value
			} else {
				field.DBName = toColumnName(fieldStruct.Name)
			}
			field.IsNormal = true
			modelStruct.StructFields = append(modelStruct.StructFields, field)
		}
	}
	modelStructsMap.Store(reflectType, &modelStruct)
	return &modelStruct
}

func parseTagSetting(tags reflect.StructTag) map[string]string {
	setting := map[string]string{}
	for _, str := range []string{tags.Get("sql"), tags.Get("zorm")} {
		tags := strings.Split(str, ";")
		for _, value := range tags {
			v := strings.Split(value, ":")
			k := strings.TrimSpace(strings.ToUpper(v[0]))
			if len(v) >= 2 {
				setting[k] = strings.Join(v[1:], ":")
			} else {
				setting[k] = k
			}
		}
	}
	return setting
}
