package main

import (
	"compress/zlib"
	"encoding/binary"
	"io"
)

type reader struct {
	data   []uint8
	offset int
}

func newReader(data []byte) *reader {
	return &reader{
		data:   data,
		offset: 0,
	}
}

func (r *reader) readUint8() (uint8, error) {
	if r.offset+1 > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}

	value := r.data[r.offset]

	r.offset += 1

	return value, nil
}

func (r *reader) readUint16() (uint16, error) {
	if r.offset+2 > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}

	value := binary.BigEndian.Uint16(r.data[r.offset:])

	r.offset += 2

	return value, nil
}

func (r *reader) readUint32() (uint32, error) {
	if r.offset+4 > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}

	value := binary.BigEndian.Uint32(r.data[r.offset:])

	r.offset += 4

	return value, nil
}

func (r *reader) readZlib() ([]byte, error) {
	z, err := zlib.NewReader(r)
	if err != nil {
		panic(err)
	}
	defer z.Close()

	data, err := io.ReadAll(z)
	if err != nil {
		panic(err)
	}

	return data, err
}

func (r *reader) readString() (string, error) {
	data, err := r.readByteSlice()
	if err != nil {
		return "", err
	}

	return string(data), err
}

func (r *reader) readByteSlice() ([]byte, error) {
	length, err := r.readUint16()
	if err != nil {
		return nil, err
	}

	newOffset := r.offset + int(length)

	if newOffset > len(r.data) {
		return nil, io.ErrUnexpectedEOF
	}

	data := r.data[r.offset:newOffset]

	r.offset = newOffset

	return data, nil
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.offset > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}

	n = copy(p, r.data[r.offset:])
	r.offset += n
	return
}

func (r *reader) ReadByte() (byte, error) {
	return r.readUint8()
}
