package bsky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bluesky-social/indigo/atproto/syntax"
)

type loginResponse struct {
	AccessJwt string `json:"accessJwt"`
	Did       string `json:"did"`
	// There are more fields, but we don't care about those.
}

type client struct {
	accessJwt, did string
}

func Login(username, password string) (*client, error) {
	req, err := createRequest("com.atproto.server.createSession", map[string]string{
		"identifier": username,
		"password":   password,
	})
	if err != nil {
		return nil, fmt.Errorf("bsky: login request: %w", err)
	}

	csr, err := sendRequest[loginResponse](req)
	if err != nil {
		return nil, fmt.Errorf("bsky: send login request: %w", err)
	}

	client := &client{
		accessJwt: csr.AccessJwt,
		did:       csr.Did,
	}

	return client, nil
}

func (c *client) Post(summary, telegramUrl string) error {
	const anchorText = "👉 Bekijk het volledige energiebericht op Telegram"

	text := fmt.Sprintf("%s\n\n%s", summary, anchorText)

	t := time.Now().UTC()
	rkey := syntax.NewTID(t.UnixMicro(), 42)

	req, err := createRequest("com.atproto.repo.applyWrites", map[string]any{
		"repo": c.did,
		"writes": []map[string]any{
			{
				"$type":      "com.atproto.repo.applyWrites#create",
				"collection": "app.bsky.feed.post",
				"rkey":       rkey,
				"value": map[string]any{
					"$type":     "app.bsky.feed.post",
					"text":      text,
					"createdAt": t,
					"langs":     []string{"nl"},
					"facets": []map[string]any{
						{
							"index": map[string]int{
								"byteStart": len(text) - len(anchorText),
								"byteEnd":   len(text),
							},
							"features": []map[string]string{
								{
									"$type": "app.bsky.richtext.facet#link",
									"uri":   telegramUrl,
								},
							},
						},
					},
				},
			},
			{
				"$type":      "com.atproto.repo.applyWrites#create",
				"collection": "app.bsky.feed.threadgate",
				"rkey":       rkey,
				"value": map[string]any{
					"$type":     "app.bsky.feed.threadgate",
					"post":      fmt.Sprintf("at://%s/app.bsky.feed.post/%s", c.did, rkey),
					"allow":     []map[string]string{},
					"createdAt": t,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("bsky: post and threadgate request: %w", err)
	}

	if _, err := sendRequest[any](req, c.accessJwt); err != nil {
		return fmt.Errorf("bsky: send post and threadgate request: %w", err)
	}

	return nil
}

func createRequest[TRequest any](method string, data TRequest) (*http.Request, error) {
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://bsky.social/xrpc/%s", method), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func sendRequest[TResponse any](req *http.Request, accessJwt ...string) (*TResponse, error) {
	if len(accessJwt) == 1 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessJwt[0]))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var r TResponse

	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &r, nil
}
