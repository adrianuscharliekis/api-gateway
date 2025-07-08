package request

type JwtRequest struct {
	GrantType string `json:"grantType" binding:"required"`
}
