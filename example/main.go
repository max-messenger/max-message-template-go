package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	template "github.com/max-messenger/max-message-template-go"
	"github.com/max-messenger/maxbot"
)

//go:embed hello.md
var tplHello string

func main() {
	bot, err := maxbot.NewApi(os.Getenv("BOT_TOKEN"), maxbot.WithHTTPClient(&http.Client{Timeout: 25 * time.Second}))
	if err != nil {
		log.Fatal(err)
	}

	bot.Handle(maxbot.OnMessageCreated, func(c maxbot.Context) error {
		_ = sendWithTemplate(bot, c)

		fmt.Println("-->", c.Update().GetMessage().Body.Text, c.Update().MessageID)

		return nil
	})

	bot.Start()
}

func sendWithTemplate(bot *maxbot.Api, c maxbot.Context) error {
	values := template.Values{
		"user_name": c.Update().GetUser().FirstName + " " + c.Update().GetUser().LastName,
		"user_id":   c.Update().GetUser().UserID,
		"bot_name":  bot.Info.Username,
		"bot_id":    bot.Info.UserID,
		"time_now":  time.Now().Format("2006-01-02 15:04:05"),
	}
	body, attachments := template.ParseFormV2(tplHello, values)
	err := c.Send(body, maxbot.WithAttachments(attachments), maxbot.WithFormat(model.FormatMarkdown))
	if err != nil {
		return err
	}

	return nil
}
