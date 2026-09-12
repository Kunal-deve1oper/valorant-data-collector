package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

const discordAPIBase = "https://discord.com/api/v10"

type discordAttachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

type discordMessage struct {
	ID          string              `json:"id"`
	Content     string              `json:"content"`
	Attachments []discordAttachment `json:"attachments"`
	Timestamp   string              `json:"timestamp"`
}

// fetchMessagesPage fetches up to `limit` messages after the given snowflake
// (empty after = from the very start of the channel). Retries on 429.
func FetchMessagesPage(client *http.Client, token, channelID, after string, limit int) ([]discordMessage, error) {
	url := fmt.Sprintf("%s/channels/%s/messages?limit=%d", discordAPIBase, channelID, limit)
	if after != "" {
		url += "&after=" + after
	}
	for {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bot "+token)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 429 {
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		if resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("discord API %d: %s", resp.StatusCode, string(b))
		}
		var msgs []discordMessage
		err = json.NewDecoder(resp.Body).Decode(&msgs)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		return msgs, nil
	}
}

// fetchAllMessagesAfter pages through a channel's full history after the
// given watermark, sorted ascending by message ID (chronological order),
// regardless of what order the API itself returns a page in.
func FetchAllMessagesAfter(client *http.Client, token, channelID, after string) ([]discordMessage, error) {
	var all []discordMessage
	for {
		batch, err := FetchMessagesPage(client, token, channelID, after, 100)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		SortMessagesByIDAscending(batch)
		all = append(all, batch...)
		after = batch[len(batch)-1].ID
		if len(batch) < 100 {
			break
		}
	}
	return all, nil
}

func FetchLatestMessage(client *http.Client, token, channelID string) (*discordMessage, error) {
	msgs, err := FetchMessagesPage(client, token, channelID, "", 1)
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 {
		return nil, nil
	}
	return &msgs[0], nil
}

func FetchMessage(client *http.Client, token, channelID, messageID string) (*discordMessage, error) {
	url := fmt.Sprintf("%s/channels/%s/messages/%s", discordAPIBase, channelID, messageID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bot "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord API %d fetching message: %s", resp.StatusCode, string(b))
	}
	var msg discordMessage
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func DownloadAttachment(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d fetching attachment", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func PostMessageContent(client *http.Client, token, channelID, content string) (string, error) {
	body, _ := json.Marshal(map[string]string{"content": content})
	url := fmt.Sprintf("%s/channels/%s/messages", discordAPIBase, channelID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("discord API %d posting message: %s", resp.StatusCode, string(b))
	}
	var msg discordMessage
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return "", err
	}
	return msg.ID, nil
}

func EditMessageContent(client *http.Client, token, channelID, messageID, content string) error {
	body, _ := json.Marshal(map[string]string{"content": content})
	url := fmt.Sprintf("%s/channels/%s/messages/%s", discordAPIBase, channelID, messageID)
	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord API %d editing message: %s", resp.StatusCode, string(b))
	}
	return nil
}

func DeleteMessage(client *http.Client, token, channelID, messageID string) error {
	url := fmt.Sprintf("%s/channels/%s/messages/%s", discordAPIBase, channelID, messageID)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bot "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord API %d deleting message: %s", resp.StatusCode, string(b))
	}
	return nil
}

// postMessageWithAttachment uploads a single file as a new message and
// returns the new message's ID.
func PostMessageWithAttachment(client *http.Client, token, channelID, filename string, fileBytes []byte, content string) (string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	payloadBytes, _ := json.Marshal(map[string]string{"content": content})
	if err := mw.WriteField("payload_json", string(payloadBytes)); err != nil {
		return "", err
	}
	part, err := mw.CreateFormFile("files[0]", filename)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(fileBytes); err != nil {
		return "", err
	}
	if err := mw.Close(); err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/channels/%s/messages", discordAPIBase, channelID)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("discord API %d posting attachment: %s", resp.StatusCode, string(b))
	}
	var msg discordMessage
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return "", err
	}
	return msg.ID, nil
}

// sortMessagesByIDAscending sorts by snowflake numeric value (not string
// comparison, to avoid lexicographic-order bugs).
func SortMessagesByIDAscending(msgs []discordMessage) {
	for i := 1; i < len(msgs); i++ {
		for j := i; j > 0; j-- {
			a, _ := strconv.ParseUint(msgs[j-1].ID, 10, 64)
			b, _ := strconv.ParseUint(msgs[j].ID, 10, 64)
			if a > b {
				msgs[j-1], msgs[j] = msgs[j], msgs[j-1]
			} else {
				break
			}
		}
	}
}
