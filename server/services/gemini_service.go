package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// GeminiService handles interactions with the Gemini API
type GeminiService struct {
	apiKey string
}

// QuoteInput represents the quote data for prompt generation
type QuoteInput struct {
	Quote  string `json:"quote"`
	Author string `json:"author"`
	Book   string `json:"book"`
}

// GeminiRequest represents the request structure for Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents a content block in the request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of the content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	Error      *GeminiError      `json:"error,omitempty"`
}

// GeminiCandidate represents a response candidate
type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

// GeminiError represents an error from the API
type GeminiError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// NewGeminiService creates a new GeminiService instance
func NewGeminiService() *GeminiService {
	return &GeminiService{
		apiKey: os.Getenv("GEMINI_API_KEY"),
	}
}

// IsConfigured returns true if the API key is set
func (s *GeminiService) IsConfigured() bool {
	return s.apiKey != ""
}

// GenerateWritingPrompt generates writing prompts based on the provided quotes
func (s *GeminiService) GenerateWritingPrompt(quotes []QuoteInput) (string, error) {
	if !s.IsConfigured() {
		return "", errors.New("Gemini API key not configured")
	}

	if len(quotes) == 0 {
		return "", errors.New("no quotes provided")
	}

	// Build the prompt
	promptText := `Based on the following quotes, generate 4 unique writing prompts.

Quotes:
`
	for i, q := range quotes {
		promptText += fmt.Sprintf("%d. \"%s\" - %s", i+1, q.Quote, q.Author)
		if q.Book != "" {
			promptText += fmt.Sprintf(", %s", q.Book)
		}
		promptText += "\n"
	}

	promptText += `
Requirements:
- Generate exactly 4 prompts that synthesize themes from these quotes
- Include 2 fiction prompts (short story, scene, character study) and 2 non-fiction prompts (personal essay, reflection, analysis)
- Each prompt should be 1-2 sentences and open-ended
- Format as a numbered list with the type in brackets, e.g. "[Fiction]" or "[Non-fiction]"
- Do NOT include any introduction, preamble, or explanation - start directly with "1."`

	// Build the request
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: promptText},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make the API request
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", s.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if geminiResp.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no response from Gemini API")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
