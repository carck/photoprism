// Package clusters provides abstract definitions of clusterers as well as
// their implementations.
package clusters

import (
	"math"
)

// DistanceFunc represents a function for measuring distance
// between n-dimensional vectors.
type DistanceFunc func([]float64, []float64) float64
type BatchDistanceFunc func(d [][]float32, ai, bi, end int) []float32

// Online represents parameters important for online learning in
// clustering algorithms.
type Online struct {
	Alpha     float64
	Dimension int
}

// HCEvent represents the intermediate result of computation of hard clustering algorithm
// and are transmitted periodically to the caller during online learning
type HCEvent struct {
	Cluster     int
	Observation []float64
}

type Clusterer32 interface {
	Learn([][]float32) error
}

// Clusterer defines the operation of learning
// common for all algorithms
type Clusterer interface {
	Learn([][]float64) error
}

// HardClusterer defines a set of operations for hard clustering algorithms
type HardClusterer interface {

	// Sizes returns sizes of respective clusters
	Sizes() []int

	// Guesses returns mapping from data point indices to cluster numbers. Clusters' numbering begins at 1.
	Guesses() []int

	// Predict returns number of cluster to which the observation would be assigned
	Predict(observation []float64) int

	// IsOnline tells the algorithm supports online learning
	IsOnline() bool

	// WithOnline configures the algorithms for online learning with given parameters
	WithOnline(Online) HardClusterer

	// Online begins the process of online training of an algorithm. Observations are sent on the observations channel,
	// once no more are expected an empty struct needs to be sent on done channel. Caller receives intermediate results of computation via
	// the returned channel.
	Online(observations chan []float64, done chan struct{}) chan *HCEvent

	// Implement common operation
	Clusterer
}

// HardClusterer defines a set of operations for hard clustering algorithms
type HardClusterer32 interface {

	// Sizes returns sizes of respective clusters
	Sizes() []int

	// Guesses returns mapping from data point indices to cluster numbers. Clusters' numbering begins at 1.
	Guesses() []int

	// Predict returns number of cluster to which the observation would be assigned
	Predict(observation []float32) int

	// IsOnline tells the algorithm supports online learning
	IsOnline() bool

	// WithOnline configures the algorithms for online learning with given parameters
	WithOnline(Online) HardClusterer32

	// Online begins the process of online training of an algorithm. Observations are sent on the observations channel,
	// once no more are expected an empty struct needs to be sent on done channel. Caller receives intermediate results of computation via
	// the returned channel.
	Online(observations chan []float32, done chan struct{}) chan *HCEvent

	// Implement common operation
	Clusterer32
}

// Estimator defines a computation used to determine an optimal number of clusters in the dataset
type Estimator interface {

	// Estimate provides an expected number of clusters in the dataset
	Estimate([][]float64) (int, error)
}

// Importer defines an operation of importing the dataset from an external file
type Importer interface {

	// Import fetches the data from a file, start and end arguments allow user
	// to specify the span of data columns to be imported (inclusively)
	Import(file string, start, end int) ([][]float64, error)
}

var (
	// EuclideanDistance is one of the common distance measurement
	EuclideanDistance = func(a, b []float64) float64 {
		var (
			s, t float64
		)

		for i := range a {
			t = a[i] - b[i]
			s += t * t
		}

		return math.Sqrt(s)
	}

	EuclideanDistance32 = func(a, b []float32) float64 {
		var (
			s, t float32
		)

		for i, _ := range a {
			t = a[i] - b[i]
			s += t * t
		}

		return math.Sqrt(float64(s))
	}

	// EuclideanDistanceSquared is one of the common distance measurement
	EuclideanDistanceSquared = func(a, b []float64) float64 {
		var (
			s, t float64
		)

		for i := range a {
			t = a[i] - b[i]
			s += t * t
		}

		return s
	}

	BatchEuclideanDistance = func(d [][]float32, ai, bi, end int) []float32 {
		var (
			s, t float32
		)
		res := make([]float32, end-bi)
		c := 0
		len := len(d[0])
		for j := bi; j < end; j++ {
			s = 0
			t = 0
			for i := 0; i < len; i++ {
				t = d[ai][i] - d[j][i]
				s += t * t
			}

			res[c] = float32(math.Sqrt(float64(s)))

			c++
		}
		return res
	}
)
