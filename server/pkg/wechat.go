package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type WechatWebhook struct {
	BaseURL string
	Key     string
	client  *http.Client
}

func NewWechatWebhook(baseURL, key string) *WechatWebhook {
	return &WechatWebhook{
		BaseURL: baseURL,
		Key:     key,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WechatWebhook) SendMarkdown(content string) error {
	if w.Key == "" {
		return fmt.Errorf("webhook key not configured")
	}

	url := fmt.Sprintf("%s?key=%s", w.BaseURL, w.Key)
	body := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := w.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("wechat webhook returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("wechat webhook error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	log.Printf("wechat webhook sent successfully")
	return nil
}

func (w *WechatWebhook) SendText(content string) error {
	if w.Key == "" {
		return fmt.Errorf("webhook key not configured")
	}

	url := fmt.Sprintf("%s?key=%s", w.BaseURL, w.Key)
	body := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := w.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("wechat webhook returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("wechat webhook error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	log.Printf("wechat webhook sent successfully")
	return nil
}
