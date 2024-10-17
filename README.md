# go722

This is a G722 codec binding with Go, source code from [sippy/libg722](https://github.com/sippy/libg722) 

## Installation
```bash
go get github.com/shenjinti/go722
```

## Usage Example

```go
func TestEncoder(t *testing.T) {
	encoder := NewG722Encoder(Rate64000, G722_DEFAULT)
	pcm := make([]byte, 160)
	g722Bytes := encoder.Encode(pcm)
	assert.NotNil(t, g722Bytes)
	assert.Equal(t, 80, len(g722Bytes))

}

func TestDecoder(t *testing.T) {
	encoder := NewG722Decoder(Rate64000, G722_DEFAULT)
	g722Bytes := make([]byte, 80)
	pcm := encoder.Decode(g722Bytes)
	assert.NotNil(t, pcm)
	assert.Equal(t, 160, len(pcm))
}
```