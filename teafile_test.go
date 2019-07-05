package goteafiles

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

type Data struct {
	Time uint64
	Price uint8
	Volume uint64
	Prob uint8
	Prib uint64
}

var data = Data{
	Time: 1299229200000,
	Price: 253,
	Volume: 8,
	Prob: 252,
	Prib: 2,
}


func TestCreate(t *testing.T) {
	tf, err := Create(
		"test.tea",
		WithDataType(reflect.TypeOf(Data{})),
		WithContentDescription("prices of acme at NYSE"),
		WithTimeFields(719162, 86400000, []int32{0}),
		WithNameValues(map[string]interface{} {
			"decimals": int32(2),
			"url"     : "www.acme.com",
		}))
	if err != nil {
		t.Fatalf("error creating TeaFile: %v", err)
	}

	err = tf.Write(data)
	if err != nil {
		t.Fatalf("error writing data to TeaFile: %v", err)
	}
	err = tf.Write(data)
	if err != nil {
		t.Fatalf("error writing data to TeaFile: %v", err)
	}

	err = tf.Close()
	if err != nil {
		t.Fatalf("error closing TeaFile: %v", err)
	}

	tf, err = OpenRead("test.tea", reflect.TypeOf(Data{}))
	if err != nil {
		t.Fatalf("error closing TeaFile: %v", err)
	}

	// Comparing with fixture
	goldTf, err := OpenRead("test-fixtures/acme.tea", reflect.TypeOf(Data{}))
	if err != nil {
		t.Fatalf("error opening golden TeaFile: %v", err)
	}

	if !reflect.DeepEqual(goldTf.itemSection, tf.itemSection) {
		t.Fatalf(
			"got different item section: %v, %v",
			goldTf.itemSection,
			tf.itemSection)
	}

	if !reflect.DeepEqual(goldTf.nameValueSection, tf.nameValueSection) {
		t.Fatalf(
			"got different item section: %v, %v",
			goldTf.nameValueSection,
			tf.nameValueSection)
	}

	if !reflect.DeepEqual(goldTf.contentDescriptionSection, tf.contentDescriptionSection) {
		t.Fatalf(
			"got different time section: %v, %v",
			goldTf.contentDescriptionSection,
			tf.contentDescriptionSection)
	}

	if !reflect.DeepEqual(goldTf.timeSection, tf.timeSection) {
		t.Fatalf(
			"got different time section: %v, %v",
			goldTf.timeSection,
			tf.timeSection)
	}
	err = os.Remove("test.tea")
	if err != nil {
		t.Fatalf("error deleting TeaFile: %v", err)
	}
}

func TestRead(t *testing.T) {
	tf, err := OpenRead("test-fixtures/acme.tea", reflect.TypeOf(Data{}))
	if err != nil {
		t.Fatalf("error opening TeaFile: %v", err)
	}
	_, err = tf.Read()
	if err != nil {
		t.Fatalf("error reading data: %v", err)
	}
}


func TestMMapRead(t *testing.T) {
	tf, err := OpenRead("test-fixtures/acme.tea", reflect.TypeOf(Data{}))
	if err != nil {
		t.Fatalf("error opening TeaFile: %v", err)
	}

	r, err := tf.OpenReadableMapping()
	if err != nil {
		t.Fatalf("error mmapping file: %v", err)
	}
	var tmp uint64
	N := r.Len()
	for i := 0; i < N; i++ {
		ptr := r.GetItem(i)
		item := (*Data)(ptr)
		tmp = item.Volume
	}
	fmt.Println(tmp)
}

func TestOBData(t *testing.T) {
	type RawOrderBookLevel struct {
		DeltaDeltaTime int16
		DeltaPrice     int32
		Quantity       uint64
	}
	tf, err := OpenRead("test-fixtures/27380280032.tea", reflect.TypeOf(RawOrderBookLevel{}))
	if err != nil {
		t.Fatalf("error opening TeaFile: %v", err)
	}

	r, err := tf.OpenReadableMapping()
	if err != nil {
		t.Fatalf("error mmapping file: %v", err)
	}
	var tmp uint64
	N := r.Len()
	for i := 0; i < N; i++ {
		ptr := r.GetItem(i)
		item := (*RawOrderBookLevel)(ptr)
		fmt.Println(item)
		tmp = item.Quantity
	}
	fmt.Println(tmp)
}