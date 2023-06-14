package types

// TelegramSendRequest 發送新訊息
type TelegramSendRequest struct {
	ChatID  string `json:"chatId"`
	Message string `json:"message"`
}

type TelegramSendResponse struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Data    TelegramSendData `json:"data"`
	Trace   string           `json:"trace"`
}
type TelegramSendData struct {
	Msg string `json:"msg"`
}

// TelegramEditRequest 編輯訊息
type TelegramEditRequest struct {
	ChatID    string `json:"chatId"`
	MessageID string `json:"messageId"`
	Message   string `json:"message"`
}

type TelegramEditResponse struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Data    TelegramEditData `json:"data"`
	Trace   string           `json:"trace"`
}

type TelegramEditData struct {
	MessageID string `json:"messageId"`
}

// TelegramDeleteRequest 刪除訊息
type TelegramDeleteRequest struct {
	ChatID    string `json:"chatId"`
	MessageID string `json:"messageId"`
}

type TelegramDeleteResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Trace   string `json:"trace"`
}
