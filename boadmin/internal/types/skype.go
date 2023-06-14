package types

// SkypeSendRequest 發送新訊息
type SkypeSendRequest struct {
	ID     string `json:"id"`
	ChatId string `json:"chatId"`
	Text   string `json:"Text"`
}

type SkypeSendResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Trace   string `json:"trace"`
}
