package mapblock

type readerCounter struct {
	inner *binaryReader
	count int64
}

func newReaderCounter(r *binaryReader) *readerCounter {
	return &readerCounter{
		inner: r,
		count: 0,
	}
}

func (r *readerCounter) Read(p []byte) (n int, err error) {
	n, err = r.inner.Read(p)
	r.count += int64(n)
	return
}

func (r *readerCounter) ReadByte() (byte, error) {
	b, err := r.inner.ReadByte()
	r.count += 1
	return b, err
}
