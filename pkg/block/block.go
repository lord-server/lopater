package block

import (
	"io"
)

type Block struct {
	Version  uint8
	NodeData []byte
}

func Decode(r io.ByteReader) (Block, error) {
	return Block{}, nil
}

func (b *Block) Encode(w io.ByteWriter) error {
	return nil
}
