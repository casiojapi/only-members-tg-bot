package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ProofRequest struct {
	Member         string `json:"member"`
	ExpectedMember string `json:"expected_member"`
}

type ProofResponse struct {
	Proof map[string]interface{} `json:"proof"`
}

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		switch update.Message.Command() {
		case "proof":
			memberID := fmt.Sprintf("%d", update.Message.From.ID)
			proof, err := requestProof(memberID, memberID)
			if err != nil {
				msg.Text = "Error generating proof: " + err.Error()
			} else {
				msg.Text = "Your proof: \n" + proof
			}
		default:
			msg.Text = "I don't know that command"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func requestProof(member string, expectedMember string) (string, error) {
	reqBody, err := json.Marshal(ProofRequest{
		Member:         member,
		ExpectedMember: expectedMember,
	})
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://localhost:8080/generate-proof", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("server error: %s", string(body))
	}

	var proofResp ProofResponse
	if err := json.NewDecoder(resp.Body).Decode(&proofResp); err != nil {
		return "", err
	}

	return convertProofToString(proofResp.Proof), nil
}

func convertProofToString(proof map[string]interface{}) string {
	var result string
	for i := 0; i < len(proof); i++ {
		if val, exists := proof[fmt.Sprintf("%d", i)]; exists {
			result += fmt.Sprintf("%v", val)
		}
	}
	return result
}
