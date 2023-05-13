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

// callbackHandle - 處理LineBot接收到的訊息，並將訊息傳遞給 Open API進行處理
// 最後將處理結果發送回Line Bot用戶。
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
			// 如果接到的訊息類型為「message」，才做訊息處裡
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					// 呼叫 getOpenAIRes 函數，根據接收到的Line Bot訊息返回對應的OpenAI 處理結果
					res, err := getOpenAIRes(message.Text)
					if err != nil {
						log.Fatal(err)
					}
					// 如果處理的結果不為空，則回覆給用戶
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

// newLineBot - 初始化LineBot
func newLineBot() (*linebot.Client, error) {
	bot, err := linebot.New(
		// LineBot的Channel Secret
		os.Getenv("CHANNEL_SECRET"),
		// LineBot的Channel Access Token
		os.Getenv("CHANNEL_ACCESS_TOKEN"),
	)

	if err != nil {
		return nil, err
	}

	return bot, nil

}

// getOpenAIRes - 取得OpenAI的回覆
func getOpenAIRes(prompt string) (string, error) {
	// 初始化 OpenAI API客戶端
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	// 設定 Completoms API 參數
	req := openai.ChatCompletionRequest{
		// 模型為3.5
		Model: openai.GPT3Dot5Turbo,
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
