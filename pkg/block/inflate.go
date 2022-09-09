package block

import (
	"bytes"
	"compress/zlib"
	"io"
)

func inflate(reader *bytes.Reader) ([]byte, error) {
	position, _ := reader.Seek(0, io.SeekCurrent)

	counter := newReaderCounter(reader)
	z, err := zlib.NewReader(counter)
	if err != nil {
		return nil, err
	}
	defer z.Close()

	data, err := io.ReadAll(z)
	if err != nil {
		return nil, err
	}

	_, err = reader.Seek(position+counter.count, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return data, err
}
