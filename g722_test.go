package go722

import (
	"testing"
)

func TestEncoder(t *testing.T) {
	encoder := NewG722Encoder(Rate64000, G722_DEFAULT)
	pcm := make([]byte, 160)
	g722Bytes := encoder.Encode(pcm)
	if g722Bytes == nil {
		t.Errorf("g722Bytes is nil")
	}
	if len(g722Bytes) != 40 {
		t.Errorf("len(g722Bytes) != 40")
	}

}

func TestDecoder(t *testing.T) {
	encoder := NewG722Decoder(Rate64000, G722_DEFAULT)
	g722Bytes := make([]byte, 40)
	pcm := encoder.Decode(g722Bytes)
	if pcm == nil {
		t.Errorf("pcm is nil")
	}
	if len(pcm) != 160 {
		t.Errorf("len(pcm) != 160")
	}
}
