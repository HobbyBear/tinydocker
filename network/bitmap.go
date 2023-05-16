package network

type bitMap struct {
	Bitmap []byte
}

func InitBitMap(maxLen int64) *bitMap {
	b := &bitMap{}
	b.Bitmap = make([]byte, maxLen)
	return b
}

func (b *bitMap) BitExist(pos int) bool {
	aIndex := arrIndex(pos)
	bIndex := bytePos(pos)
	return 1 == 1&(b.Bitmap[aIndex]>>bIndex)
}

func (b *bitMap) BitSet(pos int) {
	aIndex := arrIndex(pos)
	bIndex := bytePos(pos)
	b.Bitmap[aIndex] = b.Bitmap[aIndex] | (1 << bIndex)
}

func (b *bitMap) BitClean(pos int) {
	aIndex := arrIndex(pos)
	bIndex := bytePos(pos)
	b.Bitmap[aIndex] = b.Bitmap[aIndex] & (^(1 << bIndex))
}

func arrIndex(pos int) int {
	return pos / 8
}

func bytePos(pos int) int {
	return pos % 8
}
