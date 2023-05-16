package network

import (
	"fmt"
	"testing"
)

func TestBitSet(t *testing.T) {

	bitmap := InitBitMap(65555/8 + 1)
	fmt.Println(bitmap.BitExist(0))
	bitmap.BitSet(0)
	fmt.Println(bitmap.BitExist(0))
}
