package go722

import (
	"math"
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

type G722Band struct {
	S   int
	Sp  int
	Sz  int
	R   [3]int
	A   [3]int
	Ap  [3]int
	P   [3]int
	D   [7]int
	B   [7]int
	Bp  [7]int
	Sg  [7]int
	Nb  int
	Det int
}

type G722Encoder struct {
	// TRUE if the operating in the special ITU test mode, with the band split filters disabled.
	ItuTestMode bool
	// TRUE if the G.722 data is packed.
	Packed bool
	// TRUE if decode to 8k samples/second.
	EightK bool
	// 6 for 48000kbps, 7 for 56000kbps, or 8 for 64000kbps.
	BitsPerSample int

	// Signal history for the QMF.
	X [24]int

	Band [2]G722Band

	InBuffer  uint
	InBits    int
	OutBuffer uint
	OutBits   int
}

type G722Decoder G722Encoder

// saturate limits the amplitude to the range of int16.
func saturate(amp int32) int16 {
	amp16 := int16(amp)
	if int32(amp16) == amp {
		return amp16
	}
	if amp > math.MaxInt16 {
		return math.MaxInt16
	}
	return math.MinInt16
}

// block4 performs the Block 4 operations on the G722Band.
func block4(band *G722Band, d int) {
	var wd1, wd2, wd3 int
	var i int

	// Block 4, RECONS
	band.D[0] = d
	band.R[0] = int(saturate(int32(band.S + d)))

	// Block 4, PARREC
	band.P[0] = int(saturate(int32(band.Sz + d)))

	// Block 4, UPPOL2
	for i = 0; i < 3; i++ {
		band.Sg[i] = band.P[i] >> 15
	}
	wd1 = int(saturate(int32(band.A[1] << 2)))

	if band.Sg[0] == band.Sg[1] {
		wd2 = -wd1
	} else {
		wd2 = wd1
	}
	if wd2 > 32767 {
		wd2 = 32767
	}
	wd3 = (wd2 >> 7)
	if band.Sg[0] == band.Sg[2] {
		wd3 += 128
	} else {
		wd3 -= 128
	}
	wd3 += (band.A[2] * 32512) >> 15
	if wd3 > 12288 {
		wd3 = 12288
	} else if wd3 < -12288 {
		wd3 = -12288
	}
	band.Ap[2] = wd3

	// Block 4, UPPOL1
	band.Sg[0] = band.P[0] >> 15
	band.Sg[1] = band.P[1] >> 15
	if band.Sg[0] == band.Sg[1] {
		wd1 = 192
	} else {
		wd1 = -192
	}
	wd2 = (band.A[1] * 32640) >> 15

	band.Ap[1] = int(saturate(int32(wd1 + wd2)))
	wd3 = int(saturate(15360 - int32(band.Ap[2])))
	if band.Ap[1] > wd3 {
		band.Ap[1] = wd3
	} else if band.Ap[1] < -wd3 {
		band.Ap[1] = -wd3
	}

	// Block 4, UPZERO
	if d == 0 {
		wd1 = 0
	} else {
		wd1 = 128
	}
	band.Sg[0] = d >> 15
	for i = 1; i < 7; i++ {
		band.Sg[i] = band.D[i] >> 15
		if band.Sg[i] == band.Sg[0] {
			wd2 = wd1
		} else {
			wd2 = -wd1
		}
		wd3 = (band.B[i] * 32640) >> 15
		band.Bp[i] = int(saturate(int32(wd2 + wd3)))
	}

	// Block 4, DELAYA
	for i = 6; i > 0; i-- {
		band.D[i] = band.D[i-1]
		band.B[i] = band.Bp[i]
	}

	for i = 2; i > 0; i-- {
		band.R[i] = band.R[i-1]
		band.P[i] = band.P[i-1]
		band.A[i] = band.Ap[i]
	}

	// Block 4, FILTEP
	wd1 = int(saturate(int32(band.R[1] + band.R[1])))
	wd1 = (band.A[1] * wd1) >> 15
	wd2 = int(saturate(int32(band.R[2] + band.R[2])))
	wd2 = (band.A[2] * wd2) >> 15
	band.Sp = int(saturate(int32(wd1 + wd2)))

	// Block 4, FILTEZ
	band.Sz = 0
	for i = 6; i > 0; i-- {
		wd1 = int(saturate(int32(band.D[i] + band.D[i])))
		band.Sz += (band.B[i] * wd1) >> 15
	}
	band.Sz = int(saturate(int32(band.Sz)))

	// Block 4, PREDIC
	band.S = int(saturate(int32(band.Sp + band.Sz)))
}
