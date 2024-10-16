package clusters

import (
	"reflect"
	"testing"
)

func TestDBSCAN32Cluster(t *testing.T) {
	tests := []struct {
		MinPts   int
		Eps      float32
		Points   [][]float32
		Expected []int
	}{
		{
			MinPts:   1,
			Eps:      1,
			Points:   [][]float32{{1}},
			Expected: []int{1},
		},
		{
			MinPts:   1,
			Eps:      1,
			Points:   [][]float32{{1}, {1.5}},
			Expected: []int{1, 1},
		},
		{
			MinPts:   1,
			Eps:      1,
			Points:   [][]float32{{1}, {1}},
			Expected: []int{1, 1},
		},
		{
			MinPts:   1,
			Eps:      1,
			Points:   [][]float32{{1}, {1}, {1}},
			Expected: []int{1, 1, 1},
		},
		{
			MinPts:   1,
			Eps:      1,
			Points:   [][]float32{{1}, {1.5}, {2}},
			Expected: []int{1, 1, 1},
		},
		{
			MinPts:   1,
			Eps:      1,
			Points:   [][]float32{{1}, {1.5}, {3}},
			Expected: []int{1, 1, 2},
		},
		{
			MinPts:   2,
			Eps:      1,
			Points:   [][]float32{{1}, {3}},
			Expected: []int{-1, -1},
		},
	}
	for _, test := range tests {
		c, e := DBSCAN32(test.MinPts, test.Eps, 0, BatchEuclideanDistance)
		if e != nil {
			t.Errorf("Error initializing kmeans clusterer: %s\n", e.Error())
		}

		if e = c.Learn(test.Points); e != nil {
			t.Errorf("Error learning data: %s\n", e.Error())
		}

		if !reflect.DeepEqual(c.Guesses(), test.Expected) {
			t.Errorf("guesses does not match: %d vs %d\n", c.Guesses(), test.Expected)
		}
	}
}
