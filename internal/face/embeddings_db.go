package face

import (
	"fmt"

	"database/sql/driver"
)

func (e Embeddings) Value() (driver.Value, error) {
	if e == nil || len(e) == 0 {
		return nil, nil
	}
	return FloatsToBytes(e[0]), nil
}

func (e *Embeddings) Scan(value interface{}) error {
	if value == nil {
		*e = Embeddings{}
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
		*e = Embeddings{}
		return nil
	}

	*e = ParseEmbeddings(BytesToFloats(data))
	return nil
}
