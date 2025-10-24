package notifications

import (
	"fmt"
	"log/slog"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"log_shelter/internal/config"
)

type TelegramNotifications struct {
	bot     *tgbotapi.BotAPI
	levels  []string
	users   []int64
	enabled bool
}

type NotifyLogModel struct {
	RawLog     string  `json:"raw_log"`
	LogLevel   string  `json:"log_level"`
	Source     string  `json:"source"`
	RequestID  *string `json:"request_id,omitempty"`
	LoggerName *string `json:"logger_name,omitempty"`
}

func NewTelegramNotifications(cfg *config.TelegramConfig) (*TelegramNotifications, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.APIKey)
	if err != nil {
		return nil, err
	}

	return &TelegramNotifications{
		bot:     bot,
		levels:  cfg.Levels,
		users:   cfg.NotificateTo,
		enabled: cfg.Enabled,
	}, nil
}

func (t *TelegramNotifications) ShouldNotify(logLevel string) bool {
	return slices.Contains(t.levels, logLevel)
}

func (t *TelegramNotifications) Notify(log NotifyLogModel) {
	if !t.enabled {
		return
	}
	text := ""
	text = fmt.Sprintf("Notification from \"%v\"!\n", log.Source)
	text = fmt.Sprintf("%v[%v]\n", text, log.LogLevel)
	text = fmt.Sprintf("%vLOG:%v", text, log.RawLog)

	for _, id := range t.users {
		msg := tgbotapi.NewMessage(id, text)
		_, err := t.bot.Send(msg)
		if err != nil {
			slog.Error("Cannot notificate telegram user", "id", id, "err", err)
		}
	}
}
