package entity

// APIResponse export
type APIResponse struct {
	Success bool        `json:"success"`
	Fail    bool        `json:"fail"`
	Data    interface{} `json:"data"`
}

// NewAPIResponse export
func NewAPIResponse(data interface{}, err error) *APIResponse {
	if err != nil {
		return &APIResponse{
			Fail: true,
			Data: map[string]string{
				"reason": err.Error(),
			},
		}
	}
	return &APIResponse{
		Success: true,
		Data:    data,
	}
}
