package face

import (
	"database/sql/driver"
)

func (e Embedding) Value() (driver.Value, error) {
	if e == nil || len(e) == 0 {
		return nil, nil
	}
	return Floats32ToBytes(e), nil
}

func (e *Embedding) Scan(value interface{}) error {
	if value == nil {
		*e = face.Embedding{}
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	default:
		return fmt.Errorf("unsupported type %T for Embedding", value)
	}

	if len(data) == 0 {
		*e = face.Embedding{}
		return nil
	}

	*e = NewEmbedding(BytesToFloats32(data))
	return nil
}
