package captcha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	TurnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
)

type TurnstileVerifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIP string `json:"remoteip,omitempty"`
}

type TurnstileVerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
	Action      string   `json:"action,omitempty"`
}

type TurnstileService struct {
	secret string
}

func NewTurnstileService(secret string) *TurnstileService {
	return &TurnstileService{
		secret: secret,
	}
}

func (s *TurnstileService) VerifyToken(token, remoteIP string) (*TurnstileVerifyResponse, error) {
	if s.secret == "" {
		return nil, fmt.Errorf("TURNSTILE_SECRET not set")
	}

	reqBody := TurnstileVerifyRequest{
		Secret:   s.secret,
		Response: token,
	}
	if remoteIP != "" {
		reqBody.RemoteIP = remoteIP
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(TurnstileVerifyURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send verification request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var verifyResp TurnstileVerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &verifyResp, nil
}
