package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sashabaranov/go-openai"
	gpt3 "github.com/sashabaranov/go-openai"
)

var bot *linebot.Client

func main() {
	// 初始化 Line bot
	bot, err := newLineBot()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/collback", callbackHandle(bot))

	port := os.Getenv("PORT")

	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandle(bot *linebot.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// 解析 Line Bot 訊息
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		// 處理 Line Bot 訊息
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {

				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					res, err := getOpenAIRes(message.Text)
					if err != nil {
						log.Fatal(err)
					}

					if res != "" {
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(res)).Do(); err != nil {
							log.Print(err)

						}
					}
				}
			}
		}
	}
}
func newLineBot() (*linebot.Client, error) {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_ACCESS_TOKEN"),
	)

	if err != nil {
		return nil, err
	}

	return bot, nil

}

func getOpenAIRes(prompt string) (string, error) {
	// 初始化 OpenAI API客戶端
	client := gpt3.NewClient(os.Getenv("OPENAI_API_KEY"))

	// 設定 Completoms API 參數
	req := gpt3.ChatCompletionRequest{
		Model: gpt3.GPT3Dot5Turbo,
		// 最大輸出內容
		MaxTokens: 300,
		// 輸入的文字
		Messages: []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
		},
	}

	// 執行OpenAPI
	res, err := client.CreateChatCompletion(context.TODO(), req)
	if err != nil {
		return "", err
	}

	// 取得OpenAI回傳內容，並去除不必要的空格及分行符號
	result := strings.ReplaceAll(res.Choices[0].Message.Content, "\n", "")
	result = strings.ReplaceAll(result, " ", "")

	return result, nil

}
