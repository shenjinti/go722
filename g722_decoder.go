package go722

import "unsafe"

func NewG722Decoder(rate, options int) *G722Decoder {
	dec := &G722Decoder{}
	if rate == Rate64000 {
		dec.BitsPerSample = 6
	} else if rate == Rate56000 {
		dec.BitsPerSample = 7
	} else {
		dec.BitsPerSample = 8
	}
	if options&G722_SAMPLE_RATE_8000 != 0 {
		dec.EightK = true
	}
	if (options&G722_PACKED) != 0 && dec.BitsPerSample != 8 {
		dec.Packed = true
	}
	dec.Band[0].Det = 32
	dec.Band[1].Det = 8
	return dec
}

func (dec *G722Decoder) Decode(src []byte) (pcm []byte) {
	outbufLen := len(src)
	if !dec.EightK {
		outbufLen *= 2
	}
	outbuf := make([]int16, outbufLen)
	n := g722Decode(dec, src, len(src), outbuf)
	if n < 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&outbuf[0])), n*2)
}

// g722Decode decodes the given G.722 data to PCM samples.
func g722Decode(s *G722Decoder, g722Data []byte, length int, amp []int16) int {
	wl := [8]int{-60, -30, 58, 172, 334, 538, 1198, 3042}
	rl42 := [16]int{0, 7, 6, 5, 4, 3, 2, 1, 7, 6, 5, 4, 3, 2, 1, 0}
	ilb := [32]int{
		2048, 2093, 2139, 2186, 2233, 2282, 2332,
		2383, 2435, 2489, 2543, 2599, 2656, 2714,
		2774, 2834, 2896, 2960, 3025, 3091, 3158,
		3228, 3298, 3371, 3444, 3520, 3597, 3676,
		3756, 3838, 3922, 4008,
	}
	wh := [3]int{0, -214, 798}
	rh2 := [4]int{2, 1, 2, 1}
	qm2 := [4]int{-7408, -1616, 7408, 1616}
	qm4 := [16]int{
		0, -20456, -12896, -8968,
		-6288, -4240, -2584, -1200,
		20456, 12896, 8968, 6288,
		4240, 2584, 1200, 0,
	}
	qm5 := [32]int{
		-280, -280, -23352, -17560,
		-14120, -11664, -9752, -8184,
		-6864, -5712, -4696, -3784,
		-2960, -2208, -1520, -880,
		23352, 17560, 14120, 11664,
		9752, 8184, 6864, 5712,
		4696, 3784, 2960, 2208,
		1520, 880, 280, -280,
	}
	qm6 := [64]int{
		-136, -136, -136, -136,
		-24808, -21904, -19008, -16704,
		-14984, -13512, -12280, -11192,
		-10232, -9360, -8576, -7856,
		-7192, -6576, -6000, -5456,
		-4944, -4464, -4008, -3576,
		-3168, -2776, -2400, -2032,
		-1688, -1360, -1040, -728,
		24808, 21904, 19008, 16704,
		14984, 13512, 12280, 11192,
		10232, 9360, 8576, 7856,
		7192, 6576, 6000, 5456,
		4944, 4464, 4008, 3576,
		3168, 2776, 2400, 2032,
		1688, 1360, 1040, 728,
		432, 136, -432, -136,
	}
	qmfCoeffs := [12]int{
		3, -11, 12, 32, -210, 951, 3876, -805, 362, -156, 53, -11,
	}

	var dlowt, rlow, ihigh, dhigh, rhigh, xout1, xout2, wd1, wd2, wd3, code, outlen, i, j int

	outlen = 0
	rhigh = 0
	for j = 0; j < length; {
		if s.Packed {
			// Unpack the code bits
			if s.InBits < s.BitsPerSample {
				s.InBuffer |= uint(g722Data[j]) << s.InBits
				j++
				s.InBits += 8
			}
			code = int(s.InBuffer & ((1 << s.BitsPerSample) - 1))
			s.InBuffer >>= s.BitsPerSample
			s.InBits -= s.BitsPerSample
		} else {
			code = int(g722Data[j])
			j++
		}

		switch s.BitsPerSample {
		default:
		case 8:
			wd1 = code & 0x3F
			ihigh = (code >> 6) & 0x03
			wd2 = qm6[wd1]
			wd1 >>= 2
		case 7:
			wd1 = code & 0x1F
			ihigh = (code >> 5) & 0x03
			wd2 = qm5[wd1]
			wd1 >>= 1
		case 6:
			wd1 = code & 0x0F
			ihigh = (code >> 4) & 0x03
			wd2 = qm4[wd1]
		}
		// Block 5L, LOW BAND INVQBL
		wd2 = (s.Band[0].Det * wd2) >> 15
		// Block 5L, RECONS
		rlow = s.Band[0].S + wd2
		// Block 6L, LIMIT
		if rlow > 16383 {
			rlow = 16383
		} else if rlow < -16384 {
			rlow = -16384
		}

		// Block 2L, INVQAL
		wd2 = qm4[wd1]
		dlowt = (s.Band[0].Det * wd2) >> 15

		// Block 3L, LOGSCL
		wd2 = rl42[wd1]
		wd1 = (s.Band[0].Nb * 127) >> 7
		wd1 += wl[wd2]
		if wd1 < 0 {
			wd1 = 0
		} else if wd1 > 18432 {
			wd1 = 18432
		}
		s.Band[0].Nb = wd1

		// Block 3L, SCALEL
		wd1 = (s.Band[0].Nb >> 6) & 31
		wd2 = 8 - (s.Band[0].Nb >> 11)
		if wd2 < 0 {
			wd3 = ilb[wd1] << -wd2
		} else {
			wd3 = ilb[wd1] >> wd2
		}
		s.Band[0].Det = wd3 << 2

		block4(&s.Band[0], dlowt)

		if !s.EightK {
			// Block 2H, INVQAH
			wd2 = qm2[ihigh]
			dhigh = (s.Band[1].Det * wd2) >> 15
			// Block 5H, RECONS
			rhigh = dhigh + s.Band[1].S
			// Block 6H, LIMIT
			if rhigh > 16383 {
				rhigh = 16383
			} else if rhigh < -16384 {
				rhigh = -16384
			}

			// Block 2H, INVQAH
			wd2 = rh2[ihigh]
			wd1 = (s.Band[1].Nb * 127) >> 7
			wd1 += wh[wd2]
			if wd1 < 0 {
				wd1 = 0
			} else if wd1 > 22528 {
				wd1 = 22528
			}
			s.Band[1].Nb = wd1

			// Block 3H, SCALEH
			wd1 = (s.Band[1].Nb >> 6) & 31
			wd2 = 10 - (s.Band[1].Nb >> 11)
			if wd2 < 0 {
				wd3 = ilb[wd1] << -wd2
			} else {
				wd3 = ilb[wd1] >> wd2
			}
			s.Band[1].Det = wd3 << 2

			block4(&s.Band[1], dhigh)
		}

		if s.ItuTestMode {
			amp[outlen] = int16(rlow << 1)
			outlen++
			amp[outlen] = int16(rhigh << 1)
			outlen++
		} else {
			if s.EightK {
				amp[outlen] = int16(rlow << 1)
				outlen++
			} else {
				// Apply the receive QMF
				for i = 0; i < 22; i++ {
					s.X[i] = s.X[i+2]
				}
				s.X[22] = rlow + rhigh
				s.X[23] = rlow - rhigh

				xout1 = 0
				xout2 = 0
				for i = 0; i < 12; i++ {
					xout2 += s.X[2*i] * qmfCoeffs[i]
					xout1 += s.X[2*i+1] * qmfCoeffs[11-i]
				}
				amp[outlen] = saturate(int32(xout1 >> 11))
				outlen++
				amp[outlen] = saturate(int32(xout2 >> 11))
				outlen++
			}
		}
	}
	return outlen
}
