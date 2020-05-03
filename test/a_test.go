package test

import (
	"fmt"
	_ "fmt"
	"testing"
	"unsafe"
)

func TestA(t *testing.T) {

	bs := make([]byte, 9, 10)
	bs[1] = 67
	fmt.Println(len(bs), cap(bs))
	fmt.Println((uintptr)(unsafe.Pointer(&bs)))
	fmt.Println((*(**byte)(unsafe.Pointer(&bs))))
	fmt.Println()

	bs_ := bs[0:2]
	fmt.Println(len(bs_), cap(bs_))
	fmt.Println((uintptr)(unsafe.Pointer(&bs_)))
	fmt.Println((*(**byte)(unsafe.Pointer(&bs_))))
	// read the second
	fmt.Println(*(*byte)(unsafe.Pointer(((uintptr)(unsafe.Pointer((*(**byte)(unsafe.Pointer(&bs_))))) + uintptr(1)))))
	fmt.Println()

	bs__ := bs_[0:10]
	fmt.Println(len(bs__), cap(bs__))
	fmt.Println((uintptr)(unsafe.Pointer(&bs__)))
	fmt.Println((*(**byte)(unsafe.Pointer(&bs__))))
	fmt.Println()

	fmt.Println(bs, bs_, bs__)

	//var Len = *(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&s)) + uintptr(8)))
}
