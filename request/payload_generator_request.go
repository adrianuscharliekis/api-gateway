package request

type PayloadRequest struct {
	ClientID string `json:"client_id" binding:"required"`
	Redirect string `json:"redirect" binding:"required"`
}
