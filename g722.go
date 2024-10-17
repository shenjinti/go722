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
	ctx *C.G722_ENC_CTX
}

type G722Decoder struct {
	ctx *C.G722_DEC_CTX
}

func NewG722Encoder(rate, options int) *G722Encoder {
	ctx := C.g722_encoder_new(C.int(rate), C.int(options))
	enc := &G722Encoder{ctx: (*C.G722_ENC_CTX)(ctx)}
	runtime.SetFinalizer(enc, func(ptr any) {
		C.g722_encoder_destroy(ctx)
	})
	return enc
}

func (g *G722Encoder) Encode(pcm []byte) (dst []byte) {
	outbuf := make([]byte, len(pcm))
	n := C.g722_encode(unsafe.Pointer(g.ctx),
		(*C.int16_t)(C.CBytes(pcm)),
		C.int(len(pcm)),
		(*C.uint8_t)(C.CBytes(outbuf)))
	if n < 0 {
		return nil
	}
	return outbuf[:n]
}

func NewG722Decoder(rate, options int) *G722Decoder {
	ctx := C.g722_decoder_new(C.int(rate), C.int(options))
	dec := &G722Decoder{ctx: (*C.G722_DEC_CTX)(ctx)}
	runtime.SetFinalizer(dec, func(ptr any) {
		C.g722_decoder_destroy(ctx)
	})
	return dec
}

func (g *G722Decoder) Decode(src []byte) (pcm []byte) {
	outbuf := make([]byte, len(src)*2) // Assuming the decoded output is larger
	n := C.g722_decode(unsafe.Pointer(g.ctx),
		(*C.uint8_t)(C.CBytes(src)),
		C.int(len(src)),
		(*C.int16_t)(C.CBytes(outbuf)))
	if n < 0 {
		return nil
	}
	return outbuf[:n]
}
