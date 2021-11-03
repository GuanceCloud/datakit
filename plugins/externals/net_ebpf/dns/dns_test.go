package dns

import (
	"context"
	"testing"
	"time"
)

func TestDNS(t *testing.T) {
	dnsRecord := NewDNSRecord()
	ctx := context.Background()
	defer ctx.Done()

	if tp, err := NewTPacketDNS(); err == nil {
		go dnsRecord.Gather(ctx, tp)
	} else {
		t.Error(err)
	}
	time.Sleep(time.Second * 5)
	// t.Error(dnsRecord.record)
}
