package face

// Faces represents a list of faces detected.
type Faces []Face

// Contains returns true if the face conflicts with existing faces.
func (faces Faces) Contains(other Face) bool {
	cropArea := other.CropArea()

	for _, f := range faces {
		if f.CropArea().OverlapPercent(cropArea) > OverlapThresholdFloor {
			return true
		}
	}

	return false
}

// Append adds a face.
func (faces *Faces) Append(f Face) {
	*faces = append(*faces, f)
}

// Count returns the number of faces detected.
func (faces Faces) Count() int {
	return len(faces)
}

// Uncertainty return the max face detection uncertainty in percent.
func (faces Faces) Uncertainty() int {
	if len(faces) < 1 {
		return 100
	}

	maxScore := 0

	for _, f := range faces {
		if f.Score > maxScore {
			maxScore = f.Score
		}
	}

	return 100 - maxScore
}
