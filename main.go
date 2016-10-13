package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"

	"fmt"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/face-manager-linebot/inferences"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}
	app := &app{bot: bot}
	http.HandleFunc(os.Getenv("CALLBACK_PATH"), app.handler)
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}

type app struct {
	bot *linebot.Client
}

func (a *app) handler(w http.ResponseWriter, r *http.Request) {
	events, err := a.bot.ParseRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, event := range events {
		if event.Type != linebot.EventTypeMessage {
			log.Printf("not message event: %v", event)
			continue
		}
		if event.Source.Type != linebot.EventSourceTypeUser {
			log.Printf("not from user: %v", event)
			continue
		}
		if message, ok := event.Message.(*linebot.TextMessage); ok {
			log.Printf("text message from %s: %v", event.Source.UserID, message.Text)
			inferences, err := inferences.BulkFetch(event.Source.UserID)
			if err != nil {
				log.Println(err)
				continue
			}
			if len(inferences) < 1 {
				log.Println("empty inferences")
				continue
			}
			ids := rand.Perm(len(inferences))
			num := 5
			if len(ids) < num {
				num = len(ids)
			}
			columns := make([]*linebot.CarouselColumn, 0, 5)
			for i := 0; i < num; i++ {
				inference := inferences[ids[i]]
				columns = append(
					columns,
					linebot.NewCarouselColumn(
						inference.Face.ImageURL,
						fmt.Sprintf("id: %d", inference.Face.ID),
						fmt.Sprintf("%s [%f]", inference.Label.Name, inference.Score),
						linebot.NewURITemplateAction(
							"ソースを見る",
							inference.Face.Photo.SourceURL,
						),
						linebot.NewPostbackTemplateAction("あってる", "data", ""),
					),
				)
			}
			_, err = a.bot.ReplyMessage(
				event.ReplyToken,
				linebot.NewTemplateMessage(
					"template message",
					linebot.NewCarouselTemplate(columns...),
				),
			).Do()
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
