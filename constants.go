package goteafiles

import (
	uuid "github.com/satori/go.uuid"
	"reflect"
)

const (
	ITEM_SECTION_ID                int32 = 0x0a
	CONTENT_DESCRIPTION_SECTION_ID int32 = 0x80
	NAME_VALUE_SECTION_ID          int32 = 0x81
	TIME_SECTION_ID                int32 = 0x40

	NAME_VALUE_INT32               int32 = 1
	NAME_VALUE_DOUBLE              int32 = 2
	NAME_VALUE_TEXT                int32 = 3
	NAME_VALUE_UUID                int32 = 4
	NAME_VALUE_UINT64              int32 = 5
)

var fieldTypeToKind = map[int32]reflect.Kind {
	1 : reflect.Int8,
	2 : reflect.Int16,
	3 : reflect.Int32,
	4 : reflect.Int64,
	5 : reflect.Uint8,
	6 : reflect.Uint16,
	7 : reflect.Uint32,
	8 : reflect.Uint64,
	9 : reflect.Float32,
	10: reflect.Float64,
}

var kindToFieldType = make(map[reflect.Kind]int32)

var typeToNameValueType = map[string]int32 {
	reflect.TypeOf(int32(1)).String()    : 1,
	reflect.TypeOf(float64(1.2)).String(): 2,
	reflect.TypeOf("dede").String()   : 3,
	reflect.TypeOf(uuid.NewV1()).String(): 4,
	reflect.TypeOf(uint64(1)).String()   : 5,
}

func init() {
	for field, kind := range fieldTypeToKind {
		kindToFieldType[kind] = field
	}
}