package goteafiles

import (
	"encoding/binary"
	"fmt"
	"github.com/melaurent/goteafiles/mmap"
	//"golang.org/x/exp/mmap"
	"os"
	"reflect"
	"unsafe"
)

var nativeEndian binary.ByteOrder

func init() {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		nativeEndian = binary.LittleEndian
	case [2]byte{0xAB, 0xCD}:
		nativeEndian = binary.BigEndian
	default:
		panic("Could not determine native endianness.")
	}
}

type Header struct {
	MagicValue   int64
	ItemStart    int64
	ItemEnd      int64
	SectionCount int64
}

type TeaFile struct {
	mode                      int
	fileName                  string
	file                      *os.File
	dataType                  reflect.Type
	header                    Header
	itemSection               *ItemSection
	nameValueSection          *NameValueSection
	timeSection               *TimeSection
	contentDescriptionSection *ContentDescriptionSection
}

func Create(fileName string, configs ...TeaFileConfig) (*TeaFile, error) {
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	tf := &TeaFile{
		mode: os.O_WRONLY,
		fileName: fileName,
		file: f,
	}
	for _, config := range configs {
		config(tf)
	}
	tf.header.SectionCount = 0
	tf.header.ItemStart = int64(reflect.TypeOf(tf.header).Size())
	tf.header.ItemEnd = 0
	tf.header.MagicValue = 0x0d0e0a0402080500

	if tf.itemSection != nil {
		err = tf.checkDataType()
		if err != nil { return nil, err }

		tf.header.SectionCount += 1
		// Section ID
		tf.header.ItemStart += 4
		// Next Section Offset
		tf.header.ItemStart += 4
		// Item Section
		tf.header.ItemStart += tf.itemSection.Size()
	}

	if tf.nameValueSection != nil {
		tf.header.SectionCount += 1
		// Section ID
		tf.header.ItemStart += 4
		// Next Section Offset
		tf.header.ItemStart += 4
		// Name Value Section
		tf.header.ItemStart += tf.nameValueSection.Size()
	}

	if tf.timeSection != nil {
		tf.header.SectionCount += 1
		// Section ID
		tf.header.ItemStart += 4
		// Next Section Offset
		tf.header.ItemStart += 4
		// Time Section
		tf.header.ItemStart += tf.timeSection.Size()
	}

	if tf.contentDescriptionSection != nil {
		tf.header.SectionCount += 1
		// Section ID
		tf.header.ItemStart += 4
		// Next Section Offset
		tf.header.ItemStart += 4
		// Content Description Section
		tf.header.ItemStart += tf.contentDescriptionSection.Size()
	}

	// Align ItemStart on 8 bytes
	paddingBytes := 8 - tf.header.ItemStart % 8

	tf.header.ItemStart += paddingBytes

	err = tf.writeHeader()
	if err != nil { return nil, err }

	return tf, nil
}

func OpenRead(fileName string, dataType reflect.Type) (*TeaFile, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	tf := &TeaFile{
		mode: os.O_RDONLY,
		fileName: fileName,
		file: f,
		dataType: dataType,
	}

	err = tf.readHeader()
	if err != nil { return nil, err }

	err = tf.checkDataType()
	if err != nil { return nil, err }

	return tf, nil
}

func OpenWrite(fileName string, dataType reflect.Type) (*TeaFile, error) {
	f, err := os.OpenFile(fileName, os.O_RDWR, 0666)
	if err != nil { return nil, err }

	tf := &TeaFile{
		mode: os.O_WRONLY,
		fileName: fileName,
		file: f,
		dataType: dataType,
	}

	err = tf.readHeader()
	if err != nil { return nil, err }

	err = tf.checkDataType()
	if err != nil { return nil, err }

	_, err = f.Seek(0, 2)
	if err != nil { return nil, err }

	return tf, nil
}

func (tf *TeaFile) OpenReadableMapping() (*mmap.MMapReader, error) {
	if tf.mode == os.O_WRONLY {
		return nil, fmt.Errorf("memory mapping in write mode not supported")
	}
	_, err := tf.file.Seek(tf.header.ItemStart, 0)
	if err != nil { return nil, err }
	_, err = tf.file.Seek(0, 0)
	if err != nil { return nil, err }

	var size int64
	if tf.header.ItemEnd == 0 {
		fi, err := tf.file.Stat()
		if err != nil { return nil, err }
		size = fi.Size() - tf.header.ItemStart
	} else {
		size = tf.header.ItemEnd - tf.header.ItemStart
	}
	if size == 0 {
		return nil, fmt.Errorf("no data")
	}
	reader, err := mmap.Open(
		tf.file,
		tf.header.ItemStart,
		size,
		int64(tf.itemSection.Info.ItemSize))
	return reader, err
}

func (tf *TeaFile) Read() (interface{}, error) {
	if tf.mode == os.O_WRONLY {
		return nil, fmt.Errorf("reading in write mode not supported")
	}
	val := reflect.New(tf.dataType)
	length := int(tf.itemSection.Info.ItemSize)
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{val.Pointer(), length, length}
	b := *(*[]byte)(unsafe.Pointer(&sl))
	_, err := tf.file.Read(b)

	return val, err
}

func (tf *TeaFile) Write(val interface{}) error {
	if tf.mode == os.O_RDONLY {
		return fmt.Errorf("writing in reading mode not supported")
	}
	if tf.itemSection == nil {
		return fmt.Errorf("this file has no item section")
	}
	if reflect.TypeOf(val) != tf.dataType {
		return fmt.Errorf("was expecting %s, got %s", tf.dataType, reflect.TypeOf(val))
	}

	vp := reflect.New(reflect.TypeOf(val))
	vp.Elem().Set(reflect.ValueOf(val))
	ptr := vp.Pointer()
	length := int(tf.itemSection.Info.ItemSize)
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{ptr, length, length}
	b := *(*[]byte)(unsafe.Pointer(&sl))
	_, err := tf.file.Write(b)

	return err
}

func (tf *TeaFile) SeekItem(idx int64) error {
	if tf.mode == os.O_WRONLY {
		return fmt.Errorf("seeking in write mode not supported")
	}
	_, err := tf.file.Seek(tf.header.ItemStart + idx * int64(tf.itemSection.Info.ItemSize), 0)
	return err
}

func (tf *TeaFile) Close() error {
	return tf.file.Close()
}

func (tf *TeaFile) readHeader() error {
	err := binary.Read(tf.file, nativeEndian, &tf.header)
	if err != nil { return err }
	if tf.header.MagicValue != 0x0d0e0a0402080500 {
		return fmt.Errorf("byteordermark mismatch")
	}

	for i := 0; i < int(tf.header.SectionCount); i++ {
		var sectionID int32
		err = binary.Read(tf.file, nativeEndian, &sectionID)
		if err != nil { return err }
		var nextSectionOffset int32
		err = binary.Read(tf.file, nativeEndian, &nextSectionOffset)
		if err != nil { return err }

		beforeSection, err := tf.file.Seek(0, 1)
		if err != nil { return err }

		switch sectionID {
		case ITEM_SECTION_ID:
			tf.itemSection = &ItemSection{}
			err = tf.itemSection.Read(tf.file, nativeEndian)
			if err != nil { return err }

		case CONTENT_DESCRIPTION_SECTION_ID:
			tf.contentDescriptionSection = &ContentDescriptionSection{}
			err = tf.contentDescriptionSection.Read(tf.file, nativeEndian)
			if err != nil { return err }

		case NAME_VALUE_SECTION_ID:
			tf.nameValueSection = &NameValueSection{}
			err = tf.nameValueSection.Read(tf.file, nativeEndian)
			if err != nil { return err }

		case TIME_SECTION_ID:
			tf.timeSection = &TimeSection{}
			err = tf.timeSection.Read(tf.file, nativeEndian)
			if err != nil { return err }

		default:
			return fmt.Errorf("unknown section ID %d", sectionID)
		}

		afterSection, err := tf.file.Seek(0, 1)
		if err != nil { return err }
		if (afterSection - beforeSection) != int64(nextSectionOffset) {
			return fmt.Errorf("section reads too few or too many bytes")
		}
	}
	position, err := tf.file.Seek(0, 1)
	if err != nil { return err }

	bytesToSkip := tf.header.ItemStart - position
	_, err = tf.file.Seek(bytesToSkip, 1)
	if err != nil { return err }

	return nil
}

func (tf *TeaFile) writeHeader() error {
	var currOffset int32 = 0
	err := binary.Write(tf.file, nativeEndian, tf.header)
	if err != nil { return err }
	currOffset += int32(reflect.TypeOf(tf.header).Size())

	if tf.itemSection != nil {
		sectionSize := int32(tf.itemSection.Size())
		err = binary.Write(tf.file, nativeEndian, ITEM_SECTION_ID)
		if err != nil { return err }
		currOffset += 4
		err = binary.Write(tf.file, nativeEndian, sectionSize)
		if err != nil { return err }
		currOffset += 4
		err = tf.itemSection.Write(tf.file, nativeEndian)
		if err != nil { return err }
		currOffset += sectionSize
	}

	if tf.contentDescriptionSection != nil {
		sectionSize := int32(tf.contentDescriptionSection.Size())
		err = binary.Write(tf.file, nativeEndian, CONTENT_DESCRIPTION_SECTION_ID)
		if err != nil { return err }
		currOffset += 4
		err = binary.Write(tf.file, nativeEndian, sectionSize)
		if err != nil { return err }
		currOffset += 4
		err = tf.contentDescriptionSection.Write(tf.file, nativeEndian)
		if err != nil { return err }
		currOffset += sectionSize
	}

	if tf.nameValueSection != nil {
		sectionSize := int32(tf.nameValueSection.Size())
		err = binary.Write(tf.file, nativeEndian, NAME_VALUE_SECTION_ID)
		if err != nil { return err }
		currOffset += 4
		err = binary.Write(tf.file, nativeEndian, sectionSize)
		if err != nil { return err }
		currOffset += 4
		err = tf.nameValueSection.Write(tf.file, nativeEndian)
		if err != nil { return err }
		currOffset += sectionSize
	}

	if tf.timeSection != nil {
		sectionSize := int32(tf.timeSection.Size())
		err = binary.Write(tf.file, nativeEndian, TIME_SECTION_ID)
		if err != nil { return err }
		currOffset += 4
		err = binary.Write(tf.file, nativeEndian, sectionSize)
		if err != nil { return err }
		currOffset += 4
		err = tf.timeSection.Write(tf.file, nativeEndian)
		if err != nil { return err }
		currOffset += sectionSize
	}

	var paddingByte uint8 = 0
	for; int64(currOffset) != tf.header.ItemStart; {
		err = binary.Write(tf.file, nativeEndian, paddingByte)
		if err != nil { return err }
		currOffset += 1
	}

	return nil
}

// Check if the data type corresponds to the file description
func (tf *TeaFile) checkDataType() error {
	n := tf.dataType.NumField()
	var fields []reflect.StructField
	for i := 0; i < n; i++ {
		if tf.dataType.Field(i).Name != "_" {
			fields = append(fields, tf.dataType.Field(i))
		}
	}
	if len(fields) != len(tf.itemSection.Fields) {
		return fmt.Errorf("given type has %d fields, was expecting %d", n, len(tf.itemSection.Fields))
	}
	for i := 0; i < len(fields); i++ {
		dataField := fields[i]
		fileField := tf.itemSection.Fields[i]
		if dataField.Type.Kind() != fieldTypeToKind[fileField.Type] {
			return fmt.Errorf("was not expecting %v", dataField.Type)
		}
		if dataField.Offset != uintptr(fileField.Offset) {
			return fmt.Errorf(
				"got different offsets for field %d: %d %d",
				i,
				dataField.Offset,
				fileField.Offset)
		}
	}
	return nil
}