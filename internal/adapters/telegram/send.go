package telegram

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// SendPlain envía texto a los mismos destinos que las notificaciones de órdenes (TELEGRAM_CHAT_IDS / TELEGRAM_CHAT_ID).
func SendPlain(text string) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	rawIDs := os.Getenv("TELEGRAM_CHAT_IDS")
	if strings.TrimSpace(rawIDs) == "" {
		rawIDs = os.Getenv("TELEGRAM_CHAT_ID")
	}
	if token == "" || strings.TrimSpace(rawIDs) == "" {
		return fmt.Errorf("telegram vars faltantes")
	}
	apiURL := "https://api.telegram.org/bot" + token + "/sendMessage"
	var ids []string
	for _, part := range strings.Split(rawIDs, ",") {
		id := strings.TrimSpace(part)
		if id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return fmt.Errorf("telegram chat ids vacios")
	}
	var lastErr error
	for _, id := range ids {
		form := url.Values{}
		form.Set("chat_id", id)
		form.Set("text", text)
		form.Set("disable_web_page_preview", "1")
		resp, err := http.PostForm(apiURL, form)
		if err != nil {
			lastErr = err
			continue
		}
		func() {
			defer resp.Body.Close()
			if resp.StatusCode >= 300 {
				body, _ := io.ReadAll(resp.Body)
				lastErr = fmt.Errorf("telegram status %d: %s", resp.StatusCode, string(body))
			}
		}()
	}
	return lastErr
}

// SendToChat envía un mensaje a un chat concreto (respuestas al webhook).
func SendToChat(chatID, text string) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" || strings.TrimSpace(chatID) == "" {
		return fmt.Errorf("telegram token o chat_id faltante")
	}
	apiURL := "https://api.telegram.org/bot" + token + "/sendMessage"
	form := url.Values{}
	form.Set("chat_id", chatID)
	form.Set("text", text)
	form.Set("disable_web_page_preview", "1")
	resp, err := http.PostForm(apiURL, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
