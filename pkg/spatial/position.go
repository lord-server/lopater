package spatial

type BlockPosition struct {
	X, Y, Z int32
}

func (pos BlockPosition) Encode() int64 {
	return int64(pos.Z)*0x1000000 + int64(pos.Y)*0x1000 + int64(pos.X)
}
