package go722

/*
#include <stdint.h>
#include "g722_codec.h"
*/
import "C"
import (
	"runtime"
	"unsafe"
)

const (
	G722_DEFAULT          = 0x0000
	G722_SAMPLE_RATE_8000 = 0x0001
	G722_PACKED           = 0x0002
)

const (
	RateDefault = Rate64000
	Rate64000   = 64000
	Rate56000   = 56000
	Rate48000   = 48000
)

type G722Encoder struct {
	ctx     *C.G722_ENC_CTX
	options int
}

type G722Decoder struct {
	ctx     *C.G722_DEC_CTX
	options int
}

func NewG722Encoder(rate, options int) *G722Encoder {
	ctx := C.g722_encoder_new(C.int(rate), C.int(options))
	enc := &G722Encoder{ctx: (*C.G722_ENC_CTX)(ctx), options: options}
	runtime.SetFinalizer(enc, func(ptr any) {
		C.g722_encoder_destroy(ctx)
	})
	return enc
}

func (g *G722Encoder) Encode(pcm []byte) (dst []byte) {
	outbufLen := len(pcm)
	if g.options&G722_SAMPLE_RATE_8000 == 0 {
		outbufLen /= 2
	}

	outbuf := make([]byte, outbufLen)
	n := C.g722_encode(unsafe.Pointer(g.ctx),
		(*C.int16_t)(C.CBytes(pcm)),
		C.int(outbufLen),
		(*C.uint8_t)(unsafe.Pointer(&outbuf[0])))
	if n < 0 {
		return nil
	}
	return outbuf[:n]
}

func NewG722Decoder(rate, options int) *G722Decoder {
	ctx := C.g722_decoder_new(C.int(rate), C.int(options))
	dec := &G722Decoder{ctx: (*C.G722_DEC_CTX)(ctx), options: options}
	runtime.SetFinalizer(dec, func(ptr any) {
		C.g722_decoder_destroy(ctx)
	})
	return dec
}

func (g *G722Decoder) Decode(src []byte) (pcm []byte) {
	outbufLen := len(src)
	if g.options&G722_SAMPLE_RATE_8000 == 0 {
		outbufLen *= 2
	}

	outbuf := make([]int16, outbufLen)
	n := C.g722_decode(unsafe.Pointer(g.ctx),
		(*C.uint8_t)(C.CBytes(src)),
		C.int(len(src)),
		(*C.int16_t)(unsafe.Pointer(&outbuf[0])))
	if n < 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&outbuf[0])), n*2)
}
