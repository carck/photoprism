package photoprism

type IndexOptions struct {
	Path       string
	Rescan     bool
	Convert    bool
	Stack      bool
	FacesOnly  bool
	LabelsOnly bool
	OcrOnly    bool
}

func (o *IndexOptions) SkipUnchanged() bool {
	return !o.Rescan
}

func (o *IndexOptions) IsPartial() bool {
	return (o.OcrOnly || o.LabelsOnly || o.FacesOnly)
}

// IndexOptionsAll returns new index options with all options set to true.
func IndexOptionsAll() IndexOptions {
	result := IndexOptions{
		Path:       "/",
		Rescan:     true,
		Convert:    true,
		Stack:      true,
		FacesOnly:  false,
		LabelsOnly: false,
		OcrOnly:    false,
	}

	return result
}

// IndexOptionsFacesOnly returns new index options for updating faces only.
func IndexOptionsFacesOnly() IndexOptions {
	result := IndexOptions{
		Path:       "/",
		Rescan:     true,
		Convert:    true,
		Stack:      true,
		FacesOnly:  true,
		LabelsOnly: false,
		OcrOnly:    false,
	}

	return result
}

// IndexOptionsSingle returns new index options for unstacked, single files.
func IndexOptionsSingle() IndexOptions {
	result := IndexOptions{
		Path:       "/",
		Rescan:     true,
		Convert:    true,
		Stack:      false,
		FacesOnly:  false,
		LabelsOnly: false,
		OcrOnly:    false,
	}

	return result
}

// IndexOptionsNone returns new index options with all options set to false.
func IndexOptionsNone() IndexOptions {
	result := IndexOptions{}

	return result
}
