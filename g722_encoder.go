package go722

func NewG722Encoder(rate, options int) *G722Encoder {
	enc := &G722Encoder{}
	if rate == Rate48000 {
		enc.BitsPerSample = 6
	} else if rate == Rate56000 {
		enc.BitsPerSample = 7
	} else {
		enc.BitsPerSample = 8
	}
	if options&G722_SAMPLE_RATE_8000 != 0 {
		enc.EightK = true
	}
	if (options&G722_PACKED) != 0 && enc.BitsPerSample != 8 {
		enc.Packed = true
	}
	enc.Band[0].Det = 32
	enc.Band[1].Det = 8
	return enc
}

func (enc *G722Encoder) Encode(pcm []byte) (dst []byte) {
	outbufLen := len(pcm)
	if !enc.EightK {
		outbufLen /= 2
	}
	outbuf := make([]byte, outbufLen)
	//
	pcm16s := make([]int16, len(pcm)/2)
	for i := 0; i < len(pcm16s); i++ {
		pcm16s[i] = int16(pcm[i*2]) | int16(pcm[i*2+1])<<8
	}
	n := g722Encode(enc, pcm16s, len(pcm16s), outbuf)
	if n < 0 {
		return nil
	}
	return outbuf[:n]
}

// g722Encode encodes the given PCM samples to G.722 format.
func g722Encode(s *G722Encoder, amp []int16, length int, g722Data []byte) int {
	q6 := [32]int{
		0, 35, 72, 110, 150, 190, 233, 276,
		323, 370, 422, 473, 530, 587, 650, 714,
		786, 858, 940, 1023, 1121, 1219, 1339, 1458,
		1612, 1765, 1980, 2195, 2557, 2919, 0, 0,
	}
	iln := [32]int{
		0, 63, 62, 31, 30, 29, 28, 27,
		26, 25, 24, 23, 22, 21, 20, 19,
		18, 17, 16, 15, 14, 13, 12, 11,
		10, 9, 8, 7, 6, 5, 4, 0,
	}
	ilp := [32]int{
		0, 61, 60, 59, 58, 57, 56, 55,
		54, 53, 52, 51, 50, 49, 48, 47,
		46, 45, 44, 43, 42, 41, 40, 39,
		38, 37, 36, 35, 34, 33, 32, 0,
	}
	wl := [8]int{
		-60, -30, 58, 172, 334, 538, 1198, 3042,
	}
	rl42 := [16]int{
		0, 7, 6, 5, 4, 3, 2, 1, 7, 6, 5, 4, 3, 2, 1, 0,
	}
	ilb := [32]int{
		2048, 2093, 2139, 2186, 2233, 2282, 2332,
		2383, 2435, 2489, 2543, 2599, 2656, 2714,
		2774, 2834, 2896, 2960, 3025, 3091, 3158,
		3228, 3298, 3371, 3444, 3520, 3597, 3676,
		3756, 3838, 3922, 4008,
	}
	qm4 := [16]int{
		0, -20456, -12896, -8968,
		-6288, -4240, -2584, -1200,
		20456, 12896, 8968, 6288,
		4240, 2584, 1200, 0,
	}
	qm2 := [4]int{
		-7408, -1616, 7408, 1616,
	}
	qmfCoeffs := [12]int{
		3, -11, 12, 32, -210, 951, 3876, -805, 362, -156, 53, -11,
	}
	ihn := [3]int{0, 1, 0}
	ihp := [3]int{0, 3, 2}
	wh := [3]int{0, -214, 798}
	rh2 := [4]int{2, 1, 2, 1}

	var dlow, dhigh, el, wd, wd1, ril, wd2, il4, ih2, wd3, eh, mih, i, j int
	var xlow, xhigh, g722Bytes, sumeven, sumodd, ihigh, ilow, code int

	g722Bytes = 0
	xhigh = 0
	for j = 0; j < length; {
		if s.ItuTestMode {
			xlow = int(amp[j] >> 1)
			xhigh = xlow
			j++
		} else {
			if s.EightK {
				xlow = int(amp[j] >> 1)
				j++
			} else {
				// Apply the transmit QMF
				// Shuffle the buffer down
				copy(s.X[:22], s.X[2:])
				s.X[22] = int(amp[j])
				if j+1 < length {
					s.X[23] = int(amp[j+1])
				}
				j += 2

				// Discard every other QMF output
				sumeven = 0
				sumodd = 0
				for i = 0; i < 12; i++ {
					sumodd += s.X[2*i] * qmfCoeffs[i]
					sumeven += s.X[2*i+1] * qmfCoeffs[11-i]
				}
				xlow = (sumeven + sumodd) >> 14
				xhigh = (sumeven - sumodd) >> 14
			}
		}

		// Block 1L, SUBTRA
		el = int(saturate(int32(xlow - s.Band[0].S)))

		// Block 1L, QUANTL
		if el >= 0 {
			wd = el
		} else {
			wd = -(el + 1)
		}

		for i = 1; i < 30; i++ {
			wd1 = (q6[i] * s.Band[0].Det) >> 12
			if wd < wd1 {
				break
			}
		}

		if el < 0 {
			ilow = iln[i]
		} else {
			ilow = ilp[i]
		}

		// Block 2L, INVQAL
		ril = ilow >> 2
		wd2 = qm4[ril]
		dlow = (s.Band[0].Det * wd2) >> 15

		// Block 3L, LOGSCL
		il4 = rl42[ril]
		wd = (s.Band[0].Nb * 127) >> 7
		s.Band[0].Nb = wd + wl[il4]
		if s.Band[0].Nb < 0 {
			s.Band[0].Nb = 0
		} else if s.Band[0].Nb > 18432 {
			s.Band[0].Nb = 18432
		}

		// Block 3L, SCALEL
		wd1 = (s.Band[0].Nb >> 6) & 31
		wd2 = 8 - (s.Band[0].Nb >> 11)
		if wd2 < 0 {
			wd3 = ilb[wd1] << -wd2
		} else {
			wd3 = ilb[wd1] >> wd2
		}
		s.Band[0].Det = wd3 << 2

		block4(&s.Band[0], dlow)

		if s.EightK {
			// Just leave the high bits as zero
			code = (0xC0 | ilow) >> (8 - s.BitsPerSample)
		} else {
			// Block 1H, SUBTRA
			eh = int(saturate(int32(xhigh - s.Band[1].S)))

			// Block 1H, QUANTH

			if eh >= 0 {
				wd = eh
			} else {
				wd = -(eh + 1)
			}
			wd1 = (564 * s.Band[1].Det) >> 12

			if wd >= wd1 {
				mih = 2
			} else {
				mih = 1
			}

			if eh < 0 {
				ihigh = ihn[mih]
			} else {
				ihigh = ihp[mih]
			}

			// Block 2H, INVQAH
			wd2 = qm2[ihigh]
			dhigh = (s.Band[1].Det * wd2) >> 15

			// Block 3H, LOGSCH
			ih2 = rh2[ihigh]
			wd = (s.Band[1].Nb * 127) >> 7
			s.Band[1].Nb = wd + wh[ih2]
			if s.Band[1].Nb < 0 {
				s.Band[1].Nb = 0
			} else if s.Band[1].Nb > 22528 {
				s.Band[1].Nb = 22528
			}

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
			code = ((ihigh << 6) | ilow) >> (8 - s.BitsPerSample)
		}

		if s.Packed {
			// Pack the code bits
			s.OutBuffer |= uint(code << s.OutBits)
			s.OutBits += s.BitsPerSample
			if s.OutBits >= 8 {
				g722Data[g722Bytes] = byte(s.OutBuffer & 0xFF)
				g722Bytes++
				s.OutBits -= 8
				s.OutBuffer >>= 8
			}
		} else {
			g722Data[g722Bytes] = byte(code)
			g722Bytes++
		}
	}
	return g722Bytes
}
