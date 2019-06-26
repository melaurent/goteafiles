package goteafiles

import (
	"reflect"
)

type TeaFileConfig func(*TeaFile)

func WithDataType(typ reflect.Type) TeaFileConfig {
	// TODO populate time section if there is time types in the
	// data type
	return func (tf *TeaFile) {
		tf.dataType = typ
		itemSection := ItemSection{}
		itemSection.Info.ItemSize = int32(typ.Size())
		itemSection.Info.FieldCount = int32(typ.NumField())
		itemSection.Info.ItemTypeName = typ.Name()
		for i := 0; i < typ.NumField(); i++ {
			dataField := typ.Field(i)
			itemField := ItemSectionField{}
			itemField.Name = dataField.Name
			itemField.Offset = int32(dataField.Offset)
			itemField.Index = int32(i)
			itemField.Type = kindToFieldType[dataField.Type.Kind()]
			itemSection.Fields = append(itemSection.Fields, itemField)
		}
		tf.itemSection = &itemSection
	}
}

func WithContentDescription(description string) TeaFileConfig {
	return func (tf *TeaFile) {
		tf.contentDescriptionSection = &ContentDescriptionSection{
			ContentDescription: description,
		}
	}
}

func WithNameValues(nameValues map[string]interface{}) TeaFileConfig {
	return func (tf *TeaFile) {
		tf.nameValueSection = &NameValueSection{
			NameValues: nameValues,
		}
	}
}

func WithTimeFields(epoch int64, ticksPerDay int64, indexes []int32) TeaFileConfig {
	return func (tf *TeaFile) {
		var offsets []int32
		for _, idx := range indexes {
			offsets = append(offsets, tf.itemSection.Fields[idx].Offset)
		}
		tf.timeSection = &TimeSection{
			Epoch: epoch,
			TicksPerDay: ticksPerDay,
			Count: int32(len(offsets)),
			Offsets: offsets,
		}
	}
}