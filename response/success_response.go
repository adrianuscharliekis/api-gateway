package response

type SuccessResponse struct {
	ResponseCode    string            `json:"responseCode" binding:"required"`
	ResponseMessage string            `json:"responseMessage" binding:"required"`
	AdditionalInfo  map[string]string `json:"additionalInfo"`
}
