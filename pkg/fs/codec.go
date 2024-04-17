package fs

type FileCodec string

const (
	CodecAvc   FileCodec = "avc1"
	CodecHvc   FileCodec = "hvc1"
	CodecJpeg  FileCodec = "jpeg"
	CodecOther FileCodec = ""
)
