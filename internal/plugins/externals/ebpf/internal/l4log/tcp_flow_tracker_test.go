//go:build linux
// +build linux

package l4log

import "testing"

func TestTCPFlowTrackerSeparatesDirections(t *testing.T) {
	ft := newTCPFlowTracker(8, 1024)

	tx := ft.Push(directionTX, 100, 0, []byte("hello"), 1)
	rx := ft.Push(directionRX, 200, 0, []byte("world"), 2)

	if len(tx.Deliveries) != 1 || string(tx.Deliveries[0].Payload) != "hello" {
		t.Fatalf("unexpected tx deliveries: %+v", tx.Deliveries)
	}
	if len(rx.Deliveries) != 1 || string(rx.Deliveries[0].Payload) != "world" {
		t.Fatalf("unexpected rx deliveries: %+v", rx.Deliveries)
	}
}

func TestTCPFlowTrackerBuffersPerDirection(t *testing.T) {
	ft := newTCPFlowTracker(8, 1024)

	_ = ft.Push(directionTX, 100, 0, []byte("hello"), 1)
	res1 := ft.Push(directionTX, 110, 0, []byte("world"), 2)
	if !res1.Buffered {
		t.Fatalf("expected tx out-of-order buffering")
	}

	res2 := ft.Push(directionTX, 105, 0, []byte("_____"), 3)
	if len(res2.Deliveries) != 2 {
		t.Fatalf("deliveries = %d, want 2", len(res2.Deliveries))
	}
	if got := string(res2.Deliveries[0].Payload); got != "_____" {
		t.Fatalf("delivery[0] = %q, want %q", got, "_____")
	}
	if got := string(res2.Deliveries[1].Payload); got != "world" {
		t.Fatalf("delivery[1] = %q, want %q", got, "world")
	}
}
