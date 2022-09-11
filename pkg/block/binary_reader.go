package block

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
)

type binaryReader struct {
	*bytes.Reader
}

func newBinaryReader(data []byte) *binaryReader {
	return &binaryReader{bytes.NewReader(data)}
}

func (b *binaryReader) ReadUint8() (uint8, error) {
	return b.ReadByte()
}

func (b *binaryReader) ReadUint16() (uint16, error) {
	var value uint16
	err := binary.Read(b, binary.BigEndian, &value)
	return value, err
}

func (b *binaryReader) ReadUint32() (uint32, error) {
	var value uint32
	err := binary.Read(b, binary.BigEndian, &value)
	return value, err
}

func (b *binaryReader) ReadString() (string, error) {
	length, err := b.ReadUint16()
	if err != nil {
		return "", err
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(b, buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func (r *binaryReader) ReadZlib() ([]byte, error) {
	position, _ := r.Seek(0, io.SeekCurrent)

	counter := newReaderCounter(r)
	z, err := zlib.NewReader(counter)
	if err != nil {
		return nil, err
	}
	defer z.Close()

	data, err := io.ReadAll(z)
	if err != nil {
		return nil, err
	}

	_, err = r.Seek(position+counter.count, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return data, err
}
