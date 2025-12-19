package security

import (
	"encoding/json"
	"regexp"
	"strings"
)

// AuthParser handles authentication protocol logs
type AuthParser struct {
	Protocol string
}

func NewAuthParser(protocol string) *AuthParser {
	return &AuthParser{Protocol: protocol}
}

func (a *AuthParser) Parse(data []byte) (map[string]interface{}, error) {
	text := string(data)
	result := make(map[string]interface{})
	result["protocol"] = a.Protocol

	switch a.Protocol {
	case "OAuth2":
		return a.parseOAuth2(text)
	case "SAML":
		return a.parseSAML(text)
	case "Kerberos":
		return a.parseKerberos(text)
	case "RADIUS":
		return a.parseRADIUS(text)
	default:
		result["raw"] = text
	}

	return result, nil
}

func (a *AuthParser) parseOAuth2(text string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["protocol"] = "OAuth2"

	// Try JSON first
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(text), &jsonData); err == nil {
		return jsonData, nil
	}

	// Extract common OAuth2 fields
	if strings.Contains(text, "access_token") {
		result["event_type"] = "token_grant"
	} else if strings.Contains(text, "authorization_code") {
		result["event_type"] = "authorization"
	}

	return result, nil
}

func (a *AuthParser) parseSAML(text string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["protocol"] = "SAML"

	if strings.Contains(text, "AuthnRequest") {
		result["event_type"] = "authentication_request"
	} else if strings.Contains(text, "Response") {
		result["event_type"] = "authentication_response"
	}

	return result, nil
}

func (a *AuthParser) parseKerberos(text string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["protocol"] = "Kerberos"

	// Extract principal
	principalRe := regexp.MustCompile(`([a-zA-Z0-9_-]+)@([A-Z0-9.-]+)`)
	if matches := principalRe.FindStringSubmatch(text); len(matches) > 2 {
		result["principal"] = matches[1]
		result["realm"] = matches[2]
	}

	if strings.Contains(text, "TGT") {
		result["event_type"] = "ticket_granting_ticket"
	} else if strings.Contains(text, "ST") {
		result["event_type"] = "service_ticket"
	}

	return result, nil
}

func (a *AuthParser) parseRADIUS(text string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["protocol"] = "RADIUS"

	if strings.Contains(text, "Access-Request") {
		result["event_type"] = "access_request"
	} else if strings.Contains(text, "Access-Accept") {
		result["event_type"] = "access_accept"
	} else if strings.Contains(text, "Access-Reject") {
		result["event_type"] = "access_reject"
	}

	return result, nil
}
