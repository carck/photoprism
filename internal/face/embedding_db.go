package face

import (
	"database/sql/driver"
	"fmt"
)

func (e Embedding) Value() (driver.Value, error) {
	if e == nil || len(e) == 0 {
		return nil, nil
	}
	return FloatsToBytes(e), nil
}

func (e *Embedding) Scan(value interface{}) error {
	if value == nil {
		*e = Embedding{}
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
		*e = Embedding{}
		return nil
	}

	*e = NewEmbedding(BytesToFloats(data))
	return nil
}
