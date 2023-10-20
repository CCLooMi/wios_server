package test

import (
	"testing"
	"wios_server/handlers"
)

func TestNewBSet(t *testing.T) {
	bs := handlers.NewBSet(8 * 512 * 1024)
	t.Log(bs, len(bs))
	if len(bs) != 8+1 {
		t.Errorf("len(NewBSet(4097)) = %d; expected %d", len(bs), 9)
	}
}
func TestSetPositionValue(t *testing.T) {
	bs := []byte{0, 0, 0, 0}
	handlers.SetPositionValue(bs, 8)
	t.Log(bs)
}
func TestGetIStart(t *testing.T) {
	bs := handlers.NewBSet(8 * 512 * 1024)
	handlers.SetPositionValue(bs, 64)
	i := handlers.GetIStart(bs, 0)
	if i != 1 {
		t.Errorf("i = %d; expected %d", i, 1)
		return
	}
	t.Log(bs, i)
}
func TestCreateISetMap(t *testing.T) {
	bmap := make([]byte, 256)
	for i := 0; i < 256; i++ {
		bmap[i] = byte(getI(i))
	}
	t.Log(bmap)
}
func getI(i int) int {
	for j := 0; j < 8; j++ {
		if i&(1<<(7-j)) == 0 {
			return j
		}
	}
	return -1
}
