package zorm

import (
	"reflect"
	"sync"
)

var modelStructsMap sync.Map

type ModelStruct struct {
	PrimaryFields []*StructField
	StructFields  []*StructField
	ModelType     reflect.Type

	defaultTableName string
	l                sync.Mutex
}
