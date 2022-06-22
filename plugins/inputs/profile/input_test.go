package profile

import (
	"fmt"
	"testing"
	"time"
)

func TestMinHeap(t *testing.T) {
	heap := newMinHeap(16)

	fmt.Println(heap.getTop())

	tm1, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-08 15:04:06Z")
	tm2, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-06 15:04:06Z")
	// tm3, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-05 15:04:06Z")
	tm4, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-07 15:04:06Z")
	tm5, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-08 15:04:06Z")

	pb1 := &profileBase{
		profileID: "1111111111111111",
		birth:     tm1,
		point:     nil,
	}
	heap.push(pb1)

	pb2 := &profileBase{
		profileID: "222222222222222",
		birth:     tm2,
		point:     nil,
	}

	heap.push(pb2)

	pb3 := &profileBase{
		profileID: "3333333333333333",
		birth:     tm5,
		point:     nil,
	}

	heap.push(pb3)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	pb4 := &profileBase{
		profileID: "44444444444444",
		birth:     tm4,
		point:     nil,
	}

	heap.push(pb4)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	pb := heap.pop()

	fmt.Println(pb == pb2)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	heap.remove(pb3)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	heap.remove(pb1)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	heap.push(pb2)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)
}
