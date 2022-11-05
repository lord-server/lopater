package spatial

type MapBlockPosition struct {
	X, Y, Z int32
}

func (pos MapBlockPosition) Encode() int64 {
	return int64(pos.Z)*0x1000000 + int64(pos.Y)*0x1000 + int64(pos.X)
}
