package goteafiles

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"reflect"
	"testing"
)

func TestWithDataType(t *testing.T) {
	type Data struct {
		Time uint64
		Price uint8
		Volume uint64
		Prob uint8
		Prib uint64
	}

	// Test that a config creates the correct item section
	fixture := ItemSection{
		Info: ItemSectionInfo{
			ItemSize: 40,
			ItemTypeName: "Data",
			FieldCount: 5,
		},
		Fields: []ItemSectionField{
			{Index: 0, Type: 8, Offset: 0, Name: "Time"},
			{Index: 1, Type: 5, Offset: 8, Name: "Price"},
			{Index: 2, Type: 8, Offset: 16, Name: "Volume"},
			{Index: 3, Type: 5, Offset: 24, Name: "Prob"},
			{Index: 4, Type: 8, Offset: 32, Name: "Prib"},
		},
	}
	tf, err := Create(
		"test.tea",
		WithDataType(reflect.TypeOf(Data{})))
	if err != nil {
		t.Fatalf("error creating TeaFile: %v", err)
	}
	fmt.Println(tf.itemSection)
	if !reflect.DeepEqual(*tf.itemSection, fixture) {
		t.Fatalf("got different content description")
	}
}

func TestWithContentDescription(t *testing.T) {
	// Test that a config creates the correct description
	fixture := ContentDescriptionSection{
		ContentDescription: "prices of acme at NYSE",
	}
	tf, err := Create(
		"test.tea",
		WithContentDescription("prices of acme at NYSE"))
	if err != nil {
		t.Fatalf("error creating TeaFile: %v", err)
	}
	fmt.Println(tf.contentDescriptionSection)
	if !reflect.DeepEqual(*tf.contentDescriptionSection, fixture) {
		t.Fatalf("got different content description")
	}
}

func TestWithNameValues(t *testing.T) {
	// Test that a config creates the correct name value section
	id := uuid.NewV1()
	fixture := NameValueSection{
		NameValues: map[string]interface{} {
			"a": int32(1),
			"b": "c",
			"c": float64(1.2),
			"d": id,
			"e": uint64(100),
		},
	}
	tf, err := Create(
		"test.tea",
		WithNameValues(map[string]interface{}{
			"a": int32(1),
			"b": "c",
			"c": float64(1.2),
			"d": id,
			"e": uint64(100),
		}))
	if err != nil {
		t.Fatalf("error creating TeaFile: %v", err)
	}
	if !reflect.DeepEqual(*tf.nameValueSection, fixture) {
		t.Fatalf("got different content description")
	}
}

func TestWithTimeFields(t *testing.T) {
	// Test that a config creates the correct time section
	fixture := TimeSection{
		Epoch: 719162,
		TicksPerDay: 86400000,
		Count: 1,
		Offsets: []int32{0},
	}
	tf, err := Create(
		"test.tea",
		WithDataType(reflect.TypeOf(Data{})),
		WithTimeFields(719162, 86400000, []int32{0}))
	if err != nil {
		t.Fatalf("error creating TeaFile: %v", err)
	}
	fmt.Println(tf.timeSection)
	if !reflect.DeepEqual(*tf.timeSection, fixture) {
		t.Fatalf("got different content description")
	}
}