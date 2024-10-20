package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

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

		switch update.Message.Command() {
		case "proof":
			handleProofCommand(bot, update)
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I don't know that command")
			bot.Send(msg)
		}
	}
}

// /proof endpoint handling
func handleProofCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Use the Telegram user's ID as the member ID
	memberID := fmt.Sprintf("%d", update.Message.From.ID)

	// Request proof from the proving server
	proof, err := requestProof(memberID, memberID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error generating proof: "+err.Error())
		bot.Send(msg)
		return
	}

	filename, err := saveProofAsJSON(proof)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error saving proof to file: "+err.Error())
		bot.Send(msg)
		return
	}
	defer os.Remove(filename) 

	// send json to user
	file := tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FilePath(filename))
	if _, err := bot.Send(file); err != nil {
		log.Panic(err)
	}
}

// proof generation
func requestProof(member string, expectedMember string) (map[string]interface{}, error) {
	
	reqBody, err := json.Marshal(ProofRequest{
		Member:         member,
		ExpectedMember: expectedMember,
	})
	
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://localhost:8080/generate-proof", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// check response
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error: %s", string(body))
	}

	var proofResp ProofResponse
	if err := json.NewDecoder(resp.Body).Decode(&proofResp); err != nil {
		return nil, err
	}

	return proofResp.Proof, nil
}


func saveProofAsJSON(proof map[string]interface{}) (string, error) {
	filename := fmt.Sprintf("only_members_proof_%d.json", time.Now().Unix())

	// Marshal the proof into JSON format
	data, err := json.MarshalIndent(proof, "", "  ")
	if err != nil {
		return "", err
	}

	// write json
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return "", err
	}

	return filename, nil
}
