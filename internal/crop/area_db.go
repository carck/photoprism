package crop

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

func (e Areas) Value() (driver.Value, error) {
	if e == nil || len(e) == 0 {
		return nil, nil
	}
	return json.Marshal(e)
}

func (e *Areas) Scan(value interface{}) error {
	if value == nil {
		*e = crop.Areas{}
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported type %T for Embedding", value)
	}

	if len(data) == 0 {
		*e = crop.Areas{}
		return nil
	}

	return json.Unmarshal(data, e)
}
