package transport

type ErrorEnvelope struct {
	Error     string `json:"error"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}
type SuccessEnvelope struct {
	Data    any `json:"data"`
	Version int `json:"version,omitempty"`
}
