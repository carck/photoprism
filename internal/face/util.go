package face

import (
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
