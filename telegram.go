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

	b.RegisterHandler(bot.HandlerTypeMessageText, "/cuit", bot.MatchTypePrefix, cuit)

	log.Printf("Telegram siap! (%s)", me.Username)

	b.Start(ctx)
}

func cuit(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	if msg.From.ID == bot_id || msg.Chat.ID != config.Chat_ID {
		return
	}

	_, cuit, adacuitan := strings.Cut(msg.Text, " ")

	if msg.ReplyToMessage != nil {
		msg = msg.ReplyToMessage
		cuit = msg.Text
		adacuitan = len(msg.Text) > 0

		if msg.From.ID == bot_id {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: msg.Chat.ID,
				Text:   "Mas, Kok saya?",
			})
			return
		}

		if update.Message.From.ID != msg.From.ID && slices.Contains(config.Protected_Users, msg.From.ID) {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: msg.Chat.ID,
				Text:   "Pengguna tersebut terlindungi",
			})
			return
		}

		if !adacuitan {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: msg.Chat.ID,
				Text:   "Jenis pesan tidak didukung",
			})
			return
		}
	}

	if !adacuitan {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    msg.Chat.ID,
			Text:      fmt.Sprintf("Mas, Lu olang jangan lebay sangat.\n*Syntax:*\n/cuit `<pesan lu disini>`\n\n*Contoh:*\n/cuit gensokyo pisang keju\n\nAtau, Lu reply message orang, Lalu kirim /cuit@%s", bot_username),
			ParseMode: models.ParseModeMarkdownV1,
		})
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
	resp, err := keluarkan(text)
	if err != nil {
		log.Println(err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Yah. Ada error.",
		})
		return
	}

	url, err := getPostURL(resp)
	if err != nil {
		log.Println(err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Yah. Ada error.",
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("%s\n\n%s", text, url),
	})
}
