package utils

import "encoding/json"

// DecodeValue декодирует значение src (может быть map[string]any, pointer, value) в dst через JSON marshal/unmarshal.
// Это универсальный helper для случаев, когда interface{}/any содержит данные после JSON unmarshal.
func DecodeValue(src any, dst any) error {
	if src == nil {
		return nil
	}

	// Маршалим исходное значение в JSON и затем анмаршалим в целевой тип
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}
