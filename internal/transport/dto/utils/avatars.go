package dto

// GetAvatarsRequest запрос для получения аватарок по списку ID
type GetAvatarsRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// GetAvatarsResponse ответ со списком аватарок
type GetAvatarsResponse struct {
	Avatars map[string]*string `json:"avatars"` // map[id]url
}

func StringMapToPointerMap(m map[string]string) map[string]*string {
	if m == nil {
		return nil
	}
	result := make(map[string]*string, len(m))
	for k, v := range m {
		if v == "" {
			result[k] = nil
		} else {
			value := v
			result[k] = &value
		}
	}
	return result
}

func PointerMapToStringMap(m map[string]*string) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		if v != nil {
			result[k] = *v
		} else {
			result[k] = ""
		}
	}
	return result
}
