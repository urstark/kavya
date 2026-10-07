package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const groqCompletionsURL = "https://api.groq.com/openai/v1/chat/completions"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type ResponseFormatArgs struct {
	ReplyText       string `json:"reply_text"`
	ReactionEmoji   string `json:"reaction_emoji"`
	StickerCategory string `json:"sticker_category"`
	SendVoiceNote   string `json:"send_voice_note"`
	SendImagePrompt string `json:"send_image_prompt"`
}

type GroqClient struct {
	apiKey      string
	model       string
	visionModel string
	httpClient  *http.Client
}

func NewGroqClient(apiKey, model, visionModel string) *GroqClient {
	return &GroqClient{
		apiKey:      apiKey,
		model:       model,
		visionModel: visionModel,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (g *GroqClient) GenerateResponse(ctx context.Context, systemPrompt string, history []Message) (*ResponseFormatArgs, error) {
	messages := make([]Message, 0, len(history)+1)
	messages = append(messages, Message{Role: "system", Content: systemPrompt})
	messages = append(messages, history...)

	toolSchema := map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "format_response",
			"description": "Formats Kavya's response text, reaction emoji, and sticker category.",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"reply_text": map[string]interface{}{
						"type":        "string",
						"description": "Short 1-line conversational text, 2-14 words, or 'no_output' if staying silent.",
					},
					"reaction_emoji": map[string]interface{}{
						"type":        "string",
						"description": "Choose from [❤️, 🥰, 😍, 🤣, 😂, 😭, 🥹, 🔥, ✨, 👀, 👍, 👏, 🙏, 👌, 💯, 💔, 🤔, 🤨, 😐, 💅,🕊,👻,🗿] or 'no_reaction'.",
					},
					"sticker_category": map[string]interface{}{
						"type":        "string",
						"description": "One of [love, laughing, sassy, cool, blushing, neutral, angry, thinking, confused, crying, sad, celebrate, shocked, agreement, greeting, dismiss, playful, secret, sleepy, no_sticker]. CRITICAL: Select 'no_sticker' 80% of the time. If you select a sticker, set reply_text to 'no_output'.",
					},
					"send_voice_note": map[string]interface{}{
						"type":        "string",
						"description": "If sending a voice note, write the exact text. Use non-verbal vocalizations (laughs, sighs, hesitations, etc) for emotion. Otherwise 'no_voice'. CRITICAL: If you use this, you MUST set reply_text to 'no_output'.",
					},
					"send_image_prompt": map[string]interface{}{
						"type":        "string",
						"description": "If you want to generate and send an image, write a short visual description here. Otherwise 'no_image'. CRITICAL: If you use this, you MUST set reply_text to 'no_output'.",
					},
				},
				"required": []string{"reply_text", "reaction_emoji", "sticker_category", "send_voice_note", "send_image_prompt"},
			},
		},
	}

	payload := map[string]interface{}{
		"model":       g.model,
		"messages":    messages,
		"tools":       []interface{}{toolSchema},
		"tool_choice": map[string]interface{}{"type": "function", "function": map[string]interface{}{"name": "format_response"}},
		"temperature": 0.65,
		"max_tokens":  512,
		"top_p":       1.0,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", groqCompletionsURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("groq API error (%d): %s", resp.StatusCode, string(respBytes))
	}

	var groqResp struct {
		Choices []struct {
			Message struct {
				Role      string     `json:"role"`
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBytes, &groqResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned by groq")
	}

	choice := groqResp.Choices[0].Message
	result := &ResponseFormatArgs{
		ReactionEmoji:   "no_reaction",
		StickerCategory: "no_sticker",
		SendVoiceNote:   "no_voice",
		SendImagePrompt: "no_image",
	}

	if len(choice.ToolCalls) > 0 {
		argsStr := choice.ToolCalls[0].Function.Arguments
		if err := json.Unmarshal([]byte(argsStr), result); err != nil {
			result.ReplyText = SanitizeText(choice.Content)
		} else {
			result.ReplyText = SanitizeText(result.ReplyText)
		}
	} else {
		result.ReplyText = SanitizeText(choice.Content)
	}

	return result, nil
}

func (g *GroqClient) DescribeImage(ctx context.Context, imageData []byte) (string, error) {
	b64Image := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:image/jpeg;base64,%s", b64Image)

	messages := []map[string]interface{}{
		{
			"role": "user",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "Describe what is visible in this image in 1 brief, natural sentence like a human friend glancing at it.",
				},
				{
					"type": "image_url",
					"image_url": map[string]string{
						"url": dataURL,
					},
				},
			},
		},
	}

	payload := map[string]interface{}{
		"model":       g.visionModel,
		"messages":    messages,
		"temperature": 0.5,
		"max_tokens":  150,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", groqCompletionsURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var groqResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBytes, &groqResp); err != nil {
		return "", err
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no vision choices returned")
	}

	return groqResp.Choices[0].Message.Content, nil
}
