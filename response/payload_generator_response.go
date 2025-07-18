package response

type SignatureResponse struct {
	ClientID   string `json:"client_id"`
	Timestamp  string `json:"timestamp"`
	Signature  string `json:"signature"`
	Link       string `json:"link"`
	ExternalID string `json:"externalId"`
}
