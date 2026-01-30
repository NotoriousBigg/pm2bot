package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/NotoriousBigg/pm2bot/pm2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Process struct {
	Name  string `json:"name"`
	PMID  int    `json:"pm_id"`
	Monit struct {
		Memory int     `json:"memory"`
		CPU    float64 `json:"cpu"`
	} `json:"monit"`
	PM2Env struct {
		Status      string `json:"status"`
		RestartTime int    `json:"restart_time"`
		Unstable    int    `json:"unstable_restarts"`
	} `json:"pm2_env"`
}

func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update, allowedChatID int64) {
	if update.CallbackQuery != nil {
		handleCallback(bot, update.CallbackQuery, allowedChatID)
		return
	}

	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	if chatID != allowedChatID {
		return
	}

	text := update.Message.Text
	args := strings.Fields(text)
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}

	switch cmd {
	case "/start", "/menu", "/help":
		showMainMenu(bot, chatID, "🤖 **PM2 Control Panel**\nSelect an action below:")

	case "/list":
		sendProcessList(bot, chatID)

	case "/startapp":
		handleStartApp(bot, chatID, args)

	default:
	}
}

func showMainMenu(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = getMainKeyboard()
	bot.Send(msg)
}

func getMainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Status List", "cmd_list"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh Menu", "cmd_menu"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Restart Process", "menu_restart"),
			tgbotapi.NewInlineKeyboardButtonData("🛑 Stop Process", "menu_stop"),
		),
	)
}

func handleCallback(bot *tgbotapi.BotAPI, q *tgbotapi.CallbackQuery, allowedChatID int64) {
	if q.Message.Chat.ID != allowedChatID {
		return
	}

	bot.Request(tgbotapi.NewCallback(q.ID, ""))

	data := q.Data
	chatID := q.Message.Chat.ID
	msgID := q.Message.MessageID

	switch {
	case data == "cmd_list":
		jsonOut, err := pm2.ListJSON()
		text := ""
		if err != nil {
			text = fmt.Sprintf("❌ Error:\n`%v`", err)
		} else {
			text = formatProcessList(jsonOut)
		}

		edit := tgbotapi.NewEditMessageText(chatID, msgID, text)
		edit.ParseMode = tgbotapi.ModeMarkdown

		backRow := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Main Menu", "cmd_menu"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "cmd_list"),
		)
		edit.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{backRow}}

		bot.Send(edit)

	case data == "cmd_menu":
		edit := tgbotapi.NewEditMessageText(chatID, msgID, "🤖 **PM2 Control Panel**")
		edit.ParseMode = tgbotapi.ModeMarkdown
		kb := getMainKeyboard()
		edit.ReplyMarkup = &kb
		bot.Send(edit)

	case data == "menu_restart":
		showProcessSelection(bot, chatID, msgID, "restart")

	case data == "menu_stop":
		showProcessSelection(bot, chatID, msgID, "stop")

	case strings.HasPrefix(data, "do_restart:"):
		name := strings.TrimPrefix(data, "do_restart:")
		out, err := pm2.Restart(name)
		sendActionResult(bot, chatID, out, err)

	case strings.HasPrefix(data, "do_stop:"):
		name := strings.TrimPrefix(data, "do_stop:")
		out, err := pm2.Stop(name)
		sendActionResult(bot, chatID, out, err)
	}
}

func formatProcessList(jsonStr string) string {
	var processes []Process
	if err := json.Unmarshal([]byte(jsonStr), &processes); err != nil {
		return "❌ Error parsing process list"
	}

	if len(processes) == 0 {
		return "⚠️ No processes running."
	}

	var sb strings.Builder
	sb.WriteString("📊 **Process Status:**\n\n")

	for _, p := range processes {
		statusIcon := "⚪"
		if p.PM2Env.Status == "online" {
			statusIcon = "🟢"
		} else if p.PM2Env.Status == "errored" {
			statusIcon = "🔴"
		}

		memMB := p.Monit.Memory / 1024 / 1024

		sb.WriteString(fmt.Sprintf("%s **%s** (ID: %d)\n", statusIcon, p.Name, p.PMID))
		sb.WriteString(fmt.Sprintf("└ `%s` | 💾 %dMB | 💻 %.1f%%\n\n",
			p.PM2Env.Status, memMB, p.Monit.CPU))
	}

	return sb.String()
}

func sendProcessList(bot *tgbotapi.BotAPI, chatID int64) {
	jsonOut, err := pm2.ListJSON()
	text := ""
	if err != nil {
		text = fmt.Sprintf("❌ Error:\n`%v`", err)
	} else {
		text = formatProcessList(jsonOut)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	bot.Send(msg)
}

func showProcessSelection(bot *tgbotapi.BotAPI, chatID int64, msgID int, action string) {
	jsonStr, err := pm2.ListJSON()
	if err != nil {
		cb := tgbotapi.NewCallbackWithAlert(fmt.Sprintf("pm2 error: %s", err.Error()), "")
		bot.Send(cb)
		return
	}

	var processes []Process
	if err := json.Unmarshal([]byte(jsonStr), &processes); err != nil {
		log.Println("JSON Parse Error:", err)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, p := range processes {
		label := fmt.Sprintf("%s (%d)", p.Name, p.PMID)
		callbackData := fmt.Sprintf("do_%s:%s", action, p.Name)

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, callbackData),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 Cancel", "cmd_menu"),
	))

	edit := tgbotapi.NewEditMessageText(chatID, msgID, fmt.Sprintf("Select process to **%s**:", strings.ToUpper(action)))
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	bot.Send(edit)
}

func handleStartApp(bot *tgbotapi.BotAPI, chatID int64, args []string) {
	if len(args) < 4 {
		bot.Send(tgbotapi.NewMessage(chatID, "Usage: `/startapp <script> <name> <interpreter> [args...]`"))
		return
	}
	opt := pm2.StartOptions{
		Script:      args[1],
		Name:        args[2],
		Interpreter: args[3],
	}
	if len(args) > 4 {
		opt.Args = args[4:]
	}
	out, err := pm2.StartWithOptions(opt)
	sendActionResult(bot, chatID, out, err)
}

func sendActionResult(bot *tgbotapi.BotAPI, chatID int64, out string, err error) {
	text := ""
	if err != nil {
		text = fmt.Sprintf("❌ **Error:**\n```\n%s\n```", err.Error())
	} else {
		if out == "" {
			out = "Done"
		}
		text = fmt.Sprintf("✅ **Success:**\n```\n%s\n```", out)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	bot.Send(msg)
	showMainMenu(bot, chatID, "🤖 **PM2 Control Panel**")
}
