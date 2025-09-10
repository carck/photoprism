package face

import (
	"encoding/binary"
	"math"
)

func L2Norm(data []float32, epsilon float64) float64 {
	var sum float64 = 0
	for _, v := range data {
		sum += math.Pow(float64(v), 2)
	}
	norm := math.Sqrt(math.Max(sum, epsilon))
	for i, v := range data {
		data[i] = float32(float64(v) / norm)
	}
	return norm
}

func L2Norm64(data []float64, epsilon float64) (float64, []float32) {
	var sum float64 = 0
	for _, v := range data {
		sum += math.Pow(v, 2)
	}
	norm := math.Sqrt(math.Max(sum, epsilon))
	result := make([]float32, len(data))
	for i, v := range data {
		result[i] = float32(v / norm)
	}
	return norm, result
}

func Max(a float32, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func Max64(a float64, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func BytesToFloats(b []byte) []float32 {
	floats := make([]float32, len(b)/4)
	for i := 0; i < len(floats); i++ {
		bits := binary.LittleEndian.Uint32(b[i*4:])
		floats[i] = math.Float32frombits(bits)
	}
	return floats
}

func FloatsToBytes(floats []float32) []byte {
	byteSlice := make([]byte, 4*len(floats))
	for i, f := range floats {
		binary.LittleEndian.PutUint32(byteSlice[i*4:], math.Float32bits(f))
	}
	return byteSlice
}
