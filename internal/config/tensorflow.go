//go:build !NOTENSORFLOW
// +build !NOTENSORFLOW

package config

import (
	"path/filepath"

)

// TensorFlowVersion returns the TenorFlow framework version.
func (c *Config) TensorFlowVersion() string {
	return "1.0.0"
}

// TensorFlowModelPath returns the TensorFlow model path.
func (c *Config) TensorFlowModelPath() string {
	return filepath.Join(c.AssetsPath(), "nasnet")
}

// NSFWModelPath returns the "not safe for work" TensorFlow model path.
func (c *Config) NSFWModelPath() string {
	return filepath.Join(c.AssetsPath(), "nsfw")
}

// FaceNetModelPath returns the FaceNet model path.
func (c *Config) FaceNetModelPath() string {
	return filepath.Join(c.AssetsPath(), "facenet")
}
