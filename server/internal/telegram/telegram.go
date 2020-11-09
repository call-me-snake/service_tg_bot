package telegram

import (
	"bytes"
	"fmt"
	"github.com/call-me-snake/service_tg_bot/server/internal/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
	"log"
	"strconv"
	"text/tabwriter"
	"time"
)

const (
	greetMessage = `Привет! Это тестовый бот для сбора статистики по подключенным устройствам.
Для вывода списка устройств нажми кнопку devices. Для подробной информации по каждому устройству,
нажми кнопку с его id номером, либо введи его id.`
	helpMessage = "Для вывода списка устройств нажми кнопку devices. Для подробной информации по каждому устройству," +
		"\nнажми кнопку с его id номером, либо введи его id."
	incorrectInputMessage  = "Поддерживается только ввод id устройства, а также команды */start*, */help*"
	markdownParseMode      = "Markdown"
	storageConnectionError = "Ошибка соединения с базой устройств"
	internalError          = "Внутренняя ошибка"
	onlineStatus           = "🙂"
	offlineStatus          = "😴"
)

var deviceKeyboard = tgbotapi.InlineKeyboardMarkup{
	InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Список устройств", "devices")),
	},
}

func StartBot(botToken string, storage model.IClientDeviceStorage, logger logrus.FieldLogger) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("telegram.StartBot: %v", err)
	}

	bot.Debug = true

	logger.Info(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {

		var msg tgbotapi.MessageConfig

		if update.CallbackQuery != nil {
			switch update.CallbackQuery.Data {
			case "devices":
				msg = getDevicesList(update, storage, logger)
			case "back":
				msg = tgbotapi.NewMessage(int64(update.CallbackQuery.From.ID), "back")
			default:
				id, err := strconv.Atoi(update.CallbackQuery.Data)
				if err != nil {
					logger.Error(fmt.Sprintf("StartBot: invalid callbackQuery.Data: %s", update.CallbackQuery.Data))
					msg = tgbotapi.NewMessage(int64(update.CallbackQuery.From.ID), internalError)
				} else {
					msg = getDeviceInfoById(id, int64(update.CallbackQuery.From.ID), storage)
				}
			}
		} else if update.Message != nil {
			switch update.Message.Text {
			case "/start":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, greetMessage)
				msg.ReplyMarkup = deviceKeyboard
			case "/help":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, helpMessage)
				msg.ReplyMarkup = deviceKeyboard
			default:
				id, err := strconv.Atoi(update.Message.Text)
				if err != nil {
					log.Printf("StartBot: ")
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, incorrectInputMessage)
					msg.ParseMode = markdownParseMode
				} else {
					msg = getDeviceInfoById(id, update.Message.Chat.ID, storage)
				}
			}
		}

		_, err = bot.Send(msg)
		if err != nil {
			logrus.Error(fmt.Errorf("StartBot: %v", err))
		}
	}

	_, err = bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{})
	if err != nil {
		logger.Error(fmt.Errorf("StartBot: %v", err))
	}
	return nil
}

func getDevicesList(update tgbotapi.Update, storage model.IClientDeviceStorage, logger logrus.FieldLogger) (msg tgbotapi.MessageConfig) {
	fmt.Println("Заход в функцию getDevicesList")
	devices, err := storage.GetDeviceInfoList()
	if err != nil {
		msg = tgbotapi.NewMessage(int64(update.CallbackQuery.From.ID), storageConnectionError)
		return
	}
	if devices == nil {
		msg = tgbotapi.NewMessage(int64(update.CallbackQuery.From.ID), "Нет подключенных устройств")
		return
	}
	//Создаем инлайн клавиатуру (будут показаны первые 99 устройств + кнопка обновления)
	idKeyboard := tgbotapi.InlineKeyboardMarkup{}
	idRowCap := 8
	idRow := make([]tgbotapi.InlineKeyboardButton, 0, idRowCap)

	//Создаем строку сообщения
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 10, 0, 1, ' ', tabwriter.AlignRight)
	_, err = fmt.Fprintln(w, "```\nСписок устройств:\nID\tИмя\tСтатус\t")
	if err != nil {
		log.Print(err)
		msg = tgbotapi.NewMessage(int64(update.CallbackQuery.From.ID), "Внутренняя ошибка")
	}
	for i, device := range devices {
		if i < 99 {
			idButton := tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(int(device.Id)), strconv.Itoa(int(device.Id)))
			if len(idRow) == idRowCap {
				idKeyboard.InlineKeyboard = append(idKeyboard.InlineKeyboard, idRow)
				idRow = make([]tgbotapi.InlineKeyboardButton, 0, idRowCap)
			}
			idRow = append(idRow, idButton)
		}

		var statusMsg string
		if device.Online {
			statusMsg = onlineStatus
		} else {
			statusMsg = offlineStatus
		}
		_, err = fmt.Fprintf(w, "%d\t%s\t%s\t\n", device.Id, device.Name, statusMsg)
		if err != nil {
			logger.Error(fmt.Errorf("getDevicesList: %v", err))
		}
	}
	_, err = fmt.Fprintln(w, "```")
	if err != nil {
		logger.Error(fmt.Errorf("getDevicesList: %v", err))
	}
	err = w.Flush()
	if err != nil {
		logger.Error(fmt.Errorf("getDevicesList: %v", err))
	}

	if len(idRow) > 0 {
		idKeyboard.InlineKeyboard = append(idKeyboard.InlineKeyboard, idRow)
	}
	//Кнопка обновления списка
	idKeyboard.InlineKeyboard = append(idKeyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Обновить список устройств", "devices"),
	})
	s := string(buf.Bytes())
	fmt.Println(s)
	msg = tgbotapi.NewMessage(int64(update.CallbackQuery.From.ID), s)
	msg.ParseMode = markdownParseMode
	msg.ReplyMarkup = idKeyboard
	return
}

func getDeviceInfoById(deviceId int, userId int64, storage model.IClientDeviceStorage) (msg tgbotapi.MessageConfig) {
	device, err := storage.GetDeviceInfo(int64(deviceId))
	if err != nil {
		log.Printf("getDeviceInfoById: %v", err)
		msg = tgbotapi.NewMessage(userId, storageConnectionError)
	} else if device == nil {
		msg = tgbotapi.NewMessage(userId, fmt.Sprintf("Устройства с id: %d не найдено.", deviceId))

	} else {
		createdTimeMsg := device.CreatedAt.Format(time.RFC1123)
		var statusMsg string
		if device.Online {
			statusMsg = onlineStatus
		} else {
			statusMsg = offlineStatus
		}
		msg = tgbotapi.NewMessage(userId, fmt.Sprintf("Устройство %d:\nИмя: %s\nТокен: %s\nСоздан: %s\nОнлайн: %s",
			device.Id, device.Name, device.Token, createdTimeMsg, statusMsg))
	}
	msg.ReplyMarkup = deviceKeyboard
	return
}
