package util

import (
	"bytes"
	"encoding/binary"
	"io"
	"unicode/utf16"
)

func Read[T any](r *bytes.Reader) (tmp T) {
	binary.Read(r, binary.LittleEndian, &tmp)
	return
}

func ReadNBytes[N int | byte | uint16 | uint32](r *bytes.Reader, n N) []byte {
	tmp := make([]byte, n)
	r.Read(tmp)
	return tmp
}

func ReadString(r *bytes.Reader) string {
	r.Seek(1, io.SeekCurrent)
	tmp := make([]uint16, Read[int32](r))
	binary.Read(r, binary.LittleEndian, &tmp)
	return string(utf16.Decode(tmp))
}

func ReadNullOrStringEntry(r *bytes.Reader) string {
	typ, _ := r.ReadByte()
	ReadString(r)
	if typ != 45 {
		return ReadString(r)
	}
	return ``
}
