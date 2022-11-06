package discord

import (
	"net/http"
	"strings"
)

const webhook_URL = "https://discord.com/api/webhooks/1033486233638285313/abU-rSoTEEWMPxRehZ8H0iHSM4SN3fD9jyrRpQivAV3l71-ROnFLz3lWOizLhQVXM2Xd"

func SendMessage(message string) error {
	// Create discohook style json
	payload := `{"content": "` + message + `"}`
	// Send the message
	_, err := http.Post(webhook_URL, "application/json", strings.NewReader(payload))
	if err != nil {
		return err
	}
	return nil
}
