package main

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"slices"
	"strings"
)

var bot_username string
var bot_id int64

var helptext string

func init_telegram() {
	ctx := context.Background()
	opts := []bot.Option{}

	b, err := bot.New(config.Token, opts...)
	if err != nil {
		panic(err)
	}

	me, err := b.GetMe(ctx)
	if err != nil {
		panic(err)
	}

	bot_username = me.Username
	bot_id = me.ID

	helptext = fmt.Sprintf(
		"Mas, Lu olang jangan lebay sangat.\n"+
			"*Syntax:*\n/cuit `<pesan lu disini>`\n\n"+
			"*Contoh:*\n/cuit gensokyo pisang keju\n\n"+
			"Atau, Lu reply message orang,\n"+
			"Lalu kirim /cuit@%s",
		bot.EscapeMarkdown(bot_username),
	)

	b.RegisterHandler(bot.HandlerTypeMessageText, "/cuit", bot.MatchTypePrefix, cuit)

	log.Printf("Telegram siap! (%s)", me.Username)

	b.Start(ctx)
}

func sendMessage(ctx context.Context, b *bot.Bot, chatID int64, text string) (*models.Message, error) {
	params := &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	}

	return b.SendMessage(ctx, params)
}

func sendMessage_MarkDown(ctx context.Context, b *bot.Bot, chatID int64, text string) (*models.Message, error) {
	params := &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeMarkdownV1,
	}

	return b.SendMessage(ctx, params)
}

func sendTyping(ctx context.Context, b *bot.Bot, chatID int64) {
	params := &bot.SendChatActionParams{
		ChatID: chatID,
		Action: models.ChatActionTyping,
	}

	b.SendChatAction(ctx, params)
}

func cuit(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	if msg.From.ID == bot_id || msg.Chat.ID != config.Chat_ID {
		return
	}

	sendTyping(ctx, b, msg.Chat.ID)

	_, cuit, adacuitan := strings.Cut(msg.Text, " ")

	if msg.ReplyToMessage != nil {
		msg = msg.ReplyToMessage
		cuit = msg.Text
		adacuitan = len(msg.Text) > 0

		if msg.From.ID == bot_id {
			sendMessage(ctx, b, msg.Chat.ID, "Mas, Kok saya?")
			return
		}

		if update.Message.From.ID != msg.From.ID && slices.Contains(config.Protected_Users, msg.From.ID) {
			sendMessage(ctx, b, msg.Chat.ID, "Pengguna tersebut terlindungi")
			return
		}

		if !adacuitan {
			sendMessage(ctx, b, msg.Chat.ID, "Jenis pesan tidak didukung")
			return
		}
	}

	if !adacuitan {
		sendMessage_MarkDown(ctx, b, msg.Chat.ID, helptext)
		return
	}

	user := msg.From
	lastname := user.LastName
	if len(lastname) > 0 {
		lastname = " " + lastname
	}

	username := user.Username
	if len(username) < 1 {
		username = "<no-username>"
	}

	fullname := fmt.Sprintf("%s%s (t.me/%s)", user.FirstName, lastname, username)
	text := fmt.Sprintf("\"%s\"\n\n- %s", cuit, fullname)

	go sebarkan(ctx, b, update, text)
}

func sebarkan(ctx context.Context, b *bot.Bot, update *models.Update, text string) {
	msg := update.Message
	resp, err := keluarkan(text)
	if err != nil {
		log.Println(err)
		sendMessage(ctx, b, msg.Chat.ID, "Yah. Ada error.")
		return
	}

	url, err := getPostURL(resp)
	if err != nil {
		log.Println(err)
		sendMessage(ctx, b, msg.Chat.ID,
			"Yah. Server mastodon menolak.\n\n"+
				"Kemungkinan:\n"+
				"1. Pesan terlalu panjang\n"+
				"2. Kesalahan internal di server mastodon",
		)
		return
	}

	sendMessage(ctx, b, msg.Chat.ID,
		fmt.Sprintf("%s\n\n%s", text, url),
	)
}
