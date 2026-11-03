package ldapserver

import (
	"bufio"
	"bytes"
	"io"
	"testing"
)

// oneByteReader returns at most one byte per Read, like a slow TCP stream.
type oneByteReader struct{ r io.Reader }

func (o oneByteReader) Read(p []byte) (int, error) {
	if len(p) > 1 {
		p = p[:1]
	}
	return o.r.Read(p)
}

func TestReadMessagePacketShortReads(t *testing.T) {
	// An LDAP message larger than the bufio buffer arrives in small reads.
	body := bytes.Repeat([]byte{0x04}, 5000)
	msg := append([]byte{0x30, 0x82, byte(len(body) >> 8), byte(len(body))}, body...)
	br := bufio.NewReaderSize(oneByteReader{bytes.NewReader(msg)}, 16)
	p, err := readMessagePacket(br)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(p.bytes, msg) {
		t.Fatalf("got %d bytes, want the whole %d byte message", len(p.bytes), len(msg))
	}
}
