package goteafiles

import (
	"encoding/binary"
	"io"
)

func readText(r io.Reader, order binary.ByteOrder) (string, error) {
	var length int32
	err := binary.Read(r, order, &length)
	if err != nil { return "", err }
	strBytes := make([]byte, length)
	err = binary.Read(r, order, strBytes)
	if err != nil { return "", err }
	return string(strBytes), nil
}

func writeText(w io.Writer, order binary.ByteOrder, text string) error {
	bytes := []byte(text)
	var nBytes = int32(len(bytes))
	err := binary.Write(w, order, nBytes)
	if err != nil { return err }
	return binary.Write(w, order, bytes)
}
