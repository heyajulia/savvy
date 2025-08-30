// Package discoverpds finds the Personal Data Server for a given AT Protocol identifier.
package discoverpds

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type miniDocResponse struct {
	DID        string `json:"did"`
	Handle     string `json:"handle"`
	PDS        string `json:"pds"`
	SigningKey string `json:"signing_key"`
}

type miniDocError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// PDS finds the Personal Data Server for the given AT Protocol identifier.
func PDS(identifier string) (string, error) {
	doc, err := ResolveMiniDoc(identifier)
	if err != nil {
		return "", fmt.Errorf("resolve mini doc for %q: %w", identifier, err)
	}

	return doc.PDS, nil
}

// ResolveMiniDoc resolves the mini document for the given AT Protocol identifier.
func ResolveMiniDoc(identifier string) (*miniDocResponse, error) {
	req, err := http.NewRequest(http.MethodGet, "https://slingshot.microcosm.blue/xrpc/com.bad-example.identity.resolveMiniDoc", nil)
	if err != nil {
		return nil, fmt.Errorf("resolve DID %q: %w", identifier, err)
	}

	q := req.URL.Query()
	q.Add("identifier", identifier)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("resolve DID %q: %w", identifier, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var doc miniDocResponse

		if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode DID document response: %w", err)
		}

		return &doc, nil
	case http.StatusBadRequest:
		var doc miniDocError

		if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode error response: %w", err)
		}

		return nil, fmt.Errorf("bad request: %s: %s", doc.Error, doc.Message)
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
