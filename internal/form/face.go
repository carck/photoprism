package form

import "github.com/photoprism/photoprism/pkg/deepcopier"

// Face represents a face edit form.
type Face struct {
	FaceHidden bool   `json:"Hidden"`
	SubjUID    string `json:"SubjUID"`
}

func NewFace(m interface{}) (f Face, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
