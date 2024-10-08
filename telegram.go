package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"net/http"
	"path"
	"slices"
	"strings"
)

var bot_username string
var bot_id int64

var helptext string

var fileUpload_chan = make(chan struct{}, 1)

var Max_Attachment_Size int64 = 20000000 // 20 MB

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

func sendMessage_Markdown(ctx context.Context, b *bot.Bot, chatID int64, text string) (*models.Message, error) {
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

func sendUploading(ctx context.Context, b *bot.Bot, chatID int64) {
	params := &bot.SendChatActionParams{
		ChatID: chatID,
		Action: models.ChatActionUploadDocument,
	}

	b.SendChatAction(ctx, params)
}

func getAttachment(msg *models.Message) (hasAttachment bool, isSpoiler bool, id string) {
	var fileID string

	switch true {
	case len(msg.Photo) > 0:
		var lastWidth int
		var lastHeight int

		for _, p := range msg.Photo {
			if lastWidth > p.Width || lastHeight > p.Height {
				continue
			}

			lastWidth = p.Width
			lastHeight = p.Height
			fileID = p.FileID
		}

	case msg.Video != nil:
		if msg.Video.FileSize > Max_Attachment_Size {
			break
		}
		fileID = msg.Video.FileID

	case msg.Audio != nil:
		if msg.Audio.FileSize > Max_Attachment_Size {
			break
		}
		fileID = msg.Audio.FileID

	case msg.Sticker != nil:
		if int64(msg.Sticker.FileSize) > Max_Attachment_Size {
			break
		}
		fileID = msg.Sticker.FileID

	case msg.Voice != nil:
		if msg.Voice.FileSize > Max_Attachment_Size {
			break
		}
		fileID = msg.Voice.FileID

	case msg.VideoNote != nil:
		if int64(msg.VideoNote.FileSize) > Max_Attachment_Size {
			break
		}
		fileID = msg.VideoNote.FileID

	case msg.Animation != nil:
		if msg.Animation.FileSize > Max_Attachment_Size {
			break
		}
		fileID = msg.Animation.FileID

	case msg.Document != nil:
		if msg.Document.FileSize > Max_Attachment_Size {
			break
		}
		fileID = msg.Document.FileID
	}

	return len(fileID) > 0, msg.HasMediaSpoiler, fileID
}

func getFullName(msg *models.Message) string {
	if msg.SenderChat != nil {
		chat := msg.SenderChat
		title := chat.Title
		username := chat.Username
		return fmt.Sprintf("%s (t.me/%s)", title, username)
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

	return fmt.Sprintf("%s%s (t.me/%s)", user.FirstName, lastname, username)
}

func tooLong(ctx context.Context, b *bot.Bot, update *models.Update, text string) bool {
	if isTextTooLong(text) {
		sendMessage(ctx, b, update.Message.Chat.ID, "Suara hatimu / hatinya kepanjangan bagi server mastodon mas.")
		return true
	}

	return false
}

func cuit(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	if msg.From.ID == bot_id || msg.Chat.ID != config.Chat_ID {
		return
	}

	_, cuit, adacuitan := strings.Cut(msg.Text, " ")

	if msg.ReplyToMessage != nil && !adacuitan {
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

		hasAttachment, isSpoiler, attachment := getAttachment(msg)

		if !adacuitan && !hasAttachment {
			sendMessage(ctx, b, msg.Chat.ID, "Jenis pesan tidak didukung")
			return
		} else if hasAttachment {
			fullname := getFullName(msg)
			text := fmt.Sprintf("- %s", fullname)
			if len(msg.Caption) > 0 {
				text = fmt.Sprintf("\"%s\"\n\n", msg.Caption) + text
			}

			if tooLong(ctx, b, update, text) {
				return
			}

			go sebarkan_attachment(ctx, b, update, text, attachment, isSpoiler)
			return
		}
	}

	if !adacuitan {
		sendMessage_Markdown(ctx, b, msg.Chat.ID, helptext)
		return
	}

	fullname := getFullName(msg)
	text := fmt.Sprintf("\"%s\"\n\n- %s", cuit, fullname)

	if tooLong(ctx, b, update, text) {
		return
	}

	go sebarkan_text(ctx, b, update, text)
}

func sebarkan_text(ctx context.Context, b *bot.Bot, update *models.Update, text string) {
	sendTyping(ctx, b, update.Message.Chat.ID)

	resp, err := keluarkan(text, nil, nil)
	if err != nil {
		log.Println(err)
		sendMessage(ctx, b, update.Message.Chat.ID,
			fmt.Sprintf("Yah. Ada error.\n\n%s", err),
		)
		return
	}

	sendResult(ctx, b, update, resp)
}

func sendResult(ctx context.Context, b *bot.Bot, update *models.Update, resp *http.Response) {
	url, err := getPostURL(resp)
	if err != nil {
		log.Println(err)
		sendMessage(ctx, b, update.Message.Chat.ID,
			fmt.Sprintf("Yah. Ada error dari mastodon.\n\n%s", err),
		)
		return
	}

	sendMessage_Markdown(ctx, b, update.Message.Chat.ID,
		fmt.Sprintf("Berhasil [dicuit](%s).", url),
	)
}

func sayadulu() { fileUpload_chan <- struct{}{} }
func lepaskan() { <-fileUpload_chan }

func sebarkan_attachment(ctx context.Context, b *bot.Bot, update *models.Update, text string, fileID string, spoiler bool) {
	sayadulu()
	defer lepaskan()

	sendTyping(ctx, b, update.Message.Chat.ID)

	// get url
	f, err := b.GetFile(ctx, &bot.GetFileParams{
		FileID: fileID,
	})

	if err != nil {
		log.Println(err)
		sendMessage(ctx, b, update.Message.Chat.ID, "Mas. Maafkan aku.\nAku gagal mendapatkan URL filemu :(")
		return
	}

	// download
	sendUploading(ctx, b, update.Message.Chat.ID)

	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", config.Token, f.FilePath)

	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		sendMessage(ctx, b, update.Message.Chat.ID, "Mas...\nIni...\nKok ga mau kedownload sih 🥲\nDurov kok tega amat yak 🥲")
		return
	}

	filename := path.Base(f.FilePath)

	// create form
	defer resp.Body.Close()
	mp, formBody := createForm(filename, resp.Body)

	defer formBody.Close()

	sendUploading(ctx, b, update.Message.Chat.ID)

	// upload and get media ID
	masto_resp, masto_err := masto_postMultipart(mp, formBody)
	if masto_err != nil {
		log.Println(masto_err)
		sendMessage(ctx, b, update.Message.Chat.ID, "Yah. Ada error saat upload file ke mastodon.")
		return
	}

	defer masto_resp.Body.Close()

	var status Status
	if err := json.NewDecoder(masto_resp.Body).Decode(&status); err != nil {
		log.Println(err)
		sendMessage(ctx, b, update.Message.Chat.ID, "MastoAPI yang saya tuju sedang mabuk :/\n\nKu harap MastoAPInya respon dengan JSON, Lah yang datang di luar ekspetasi gue.")
		return
	}

	if len(status.Error) > 0 {
		log.Println("Upload to Mastodon error:", status.Error)
		sendMessage(ctx, b, update.Message.Chat.ID, fmt.Sprintf("Ku sudah berusaha upload ke mastodon, Namun MastoAPI bilang:\n\n%s", status.Error))
		return
	}

	sendTyping(ctx, b, update.Message.Chat.ID)

	// media still processing. wait till ready
	if masto_resp.StatusCode == 202 {
		err := masto_wait_media_ready(status.ID)

		if err != nil {
			log.Println(err)
			sendMessage(ctx, b, update.Message.Chat.ID,
				fmt.Sprintf("Yah. Ada error.\n\n%s", err),
			)
		}
	}

	// post
	toot_resp, toot_err := keluarkan(text, &status.ID, &spoiler)
	if toot_err != nil {
		log.Println(err)
		sendMessage(ctx, b, update.Message.Chat.ID,
			fmt.Sprintf("Yah. Ada error.\n\n%s", err),
		)
		return
	}

	sendResult(ctx, b, update, toot_resp)
}
