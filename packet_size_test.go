package ldapserver

import (
	"bufio"
	"bytes"
	"runtime"
	"testing"
)

func TestReadMessagePacketRejectsHugeLength(t *testing.T) {
	// A SEQUENCE header that announces a body of 2147483647 bytes and sends none.
	header := []byte{0x30, 0x84, 0x7f, 0xff, 0xff, 0xff}

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	_, err := readMessagePacket(bufio.NewReader(bytes.NewReader(header)))
	runtime.ReadMemStats(&after)

	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 1<<20 {
		t.Errorf("reading a 6 byte header allocated %d bytes", allocated)
	}
	if err == nil {
		t.Error("expected an error for a message larger than MaxMessageSize")
	}
}

func TestReadMessagePacketAcceptsLimit(t *testing.T) {
	old := MaxMessageSize
	MaxMessageSize = len(rawBindRequest) - 2
	defer func() { MaxMessageSize = old }()

	if _, err := readMessagePacket(bufio.NewReader(bytes.NewReader(rawBindRequest))); err != nil {
		t.Fatalf("message at the limit was rejected: %v", err)
	}
	MaxMessageSize--
	if _, err := readMessagePacket(bufio.NewReader(bytes.NewReader(rawBindRequest))); err == nil {
		t.Fatal("message over the limit was accepted")
	}
}

func TestReadMessagePacketAllocatesOnlyReceivedBytes(t *testing.T) {
	// A header that announces MaxMessageSize body bytes and then ends.
	header := []byte{0x30, 0x84, byte(MaxMessageSize >> 24), byte(MaxMessageSize >> 16), byte(MaxMessageSize >> 8), byte(MaxMessageSize)}

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	_, err := readMessagePacket(bufio.NewReader(bytes.NewReader(header)))
	runtime.ReadMemStats(&after)

	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 256<<10 {
		t.Errorf("a header with no body allocated %d bytes", allocated)
	}
	if err == nil {
		t.Error("expected an error for a message with a missing body")
	}
}
