package types

import (
	"fmt"
	"strings"
)

type UUIDResponse struct {
	Uuid string `json:"uuid"`
}

type UserAgentResponse struct {
	UserAgent string `json:"user-agent"`
}

type IPResponse struct {
	Origin string `json:"origin"`
}

type FetchChannelResponse struct {
	Uuid      string `json:"uuid"`
	UserAgent string `json:"user_agent"`
	Origin    string `json:"origin"`
}

type FetchResult struct {
	Url  string
	Data any
	Err  error
}

// GetModelForURL looks at a dynamic URL and returns an empty structural pointer
func GetModelForURL(url string) (any, error) {
	switch {
	case strings.Contains(url, "/uuid"):
		return &UUIDResponse{}, nil
	case strings.Contains(url, "/user-agent"):
		return &UserAgentResponse{}, nil
	case strings.Contains(url, "/ip"):
		return &IPResponse{}, nil
	default:
		return nil, fmt.Errorf("unsupported dynamic route: %s", url)
	}
}
