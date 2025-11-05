package entity

type TelegramUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *TelegramMessage `json:"message"`
}

type TelegramMessage struct {
	MessageID int          `json:"message_id"`
	From      TelegramUser `json:"from"`
	ChatID    TelegramChat `json:"chat_id"`
	Date      int          `json:"date"`
	Text      string       `json:"text"`
}

type TelegramUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type TelegramChat struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Type     string `json:"type"`
}
