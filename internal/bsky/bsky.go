package bsky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type loginResponse struct {
	AccessJwt string `json:"accessJwt"`
	Did       string `json:"did"`
	Handle    string `json:"handle"`
	// There are more fields, but these are just the ones we care about.
}

type postRequest struct {
	Repo       string     `json:"repo"`
	Collection string     `json:"collection"`
	Record     postRecord `json:"record"`
}

type postRecord struct {
	Type      string    `json:"$type"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	Langs     []string  `json:"langs"`
	Facets    []facet   `json:"facets"`
}

type facet struct {
	Index    index     `json:"index"`
	Features []feature `json:"features"`
}

type index struct {
	ByteStart int `json:"byteStart"`
	ByteEnd   int `json:"byteEnd"`
}

type feature struct {
	Type string `json:"$type"`
	Uri  string `json:"uri"`
}

type postResponse struct {
	Uri string `json:"uri"`
}

type restrictRepliesRecord struct {
	Type      string    `json:"$type"`
	Post      string    `json:"post"`
	Allow     []any     `json:"allow"`
	CreatedAt time.Time `json:"createdAt"`
}

type restrictRepliesRequest struct {
	Rkey       string                `json:"rkey"`
	Repo       string                `json:"repo"`
	Collection string                `json:"collection"`
	Record     restrictRepliesRecord `json:"record"`
}

type client struct {
	accessJwt, did, handle string
}

func Login(username, password string) (*client, error) {
	data := loginRequest{
		Identifier: username,
		Password:   password,
	}

	req, err := createRequest("com.atproto.server.createSession", data)
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
		handle:    csr.Handle,
	}

	return client, nil
}

func (c *client) Post(summary, telegramUrl string) error {
	uri, err := c.postSummary(summary, telegramUrl)
	if err != nil {
		return fmt.Errorf("bsky: post summary: %w", err)
	}

	parts := strings.Split(uri, "/")

	// This isn't great validation, but it at least prevents our code from panicking.
	if len(parts) < 2 {
		return fmt.Errorf("bsky: malformed post uri '%s'", uri)
	}

	rkey := parts[len(parts)-1]

	if err := c.restrictReplies(uri, rkey); err != nil {
		return fmt.Errorf("bsky: restrict replies to '%s': %w", uri, err)
	}

	return nil
}

func (c *client) postSummary(summary, telegramUrl string) (string, error) {
	const anchorText = "👉 Bekijk het volledige energiebericht op Telegram"

	text := fmt.Sprintf("%s\n\n%s", summary, anchorText)
	data := postRequest{
		Repo:       c.did,
		Collection: "app.bsky.feed.post",
		Record: postRecord{
			Type:      "app.bsky.feed.post",
			Text:      text,
			CreatedAt: time.Now().UTC(),
			Langs:     []string{"nl"},
			Facets: []facet{
				{
					Index: index{
						ByteStart: len(text) - len(anchorText),
						ByteEnd:   len(text),
					},
					Features: []feature{
						{
							Type: "app.bsky.richtext.facet#link",
							Uri:  telegramUrl,
						},
					},
				},
			},
		},
	}

	req, err := createRequest("com.atproto.repo.createRecord", data)
	if err != nil {
		return "", fmt.Errorf("post summary request: %w", err)
	}

	crr, err := sendRequest[postResponse](req, c.accessJwt)
	if err != nil {
		return "", fmt.Errorf("send post summary request: %w", err)
	}

	return crr.Uri, nil
}

func (c *client) restrictReplies(uri, rkey string) error {
	data := restrictRepliesRequest{
		Rkey:       rkey,
		Repo:       c.did,
		Collection: "app.bsky.feed.threadgate",
		Record: restrictRepliesRecord{
			Type:      "app.bsky.feed.threadgate",
			Post:      uri,
			Allow:     []any{},
			CreatedAt: time.Now().UTC(),
		},
	}

	req, err := createRequest("com.atproto.repo.createRecord", data)
	if err != nil {
		return fmt.Errorf("restrict replies request: %w", err)
	}

	if _, err = sendRequest[any](req, c.accessJwt); err != nil {
		return fmt.Errorf("send restrict replies request: %w", err)
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
