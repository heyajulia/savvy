package bsky

import (
	"context"
	"fmt"
	"time"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	appbsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/atproto/syntax"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/heyajulia/savvy/pkg/discoverpds"
)

type client struct {
	client *xrpc.Client
}

func Login(username, password string) (*client, error) {
	// TODO: Handle contexts better?

	pds, err := discoverpds.PDS(username)
	if err != nil {
		return nil, fmt.Errorf("bsky: find PDS: %w", err)
	}

	xrpcc := &xrpc.Client{
		Client: util.RobustHTTPClient(),
		Host:   pds,
		Auth:   &xrpc.AuthInfo{Handle: username},
	}

	auth, err := comatproto.ServerCreateSession(context.Background(), xrpcc, &comatproto.ServerCreateSession_Input{
		Identifier: xrpcc.Auth.Handle,
		Password:   password,
	})
	if err != nil {
		return nil, fmt.Errorf("bsky: create session: %w", err)
	}

	xrpcc.Auth.AccessJwt = auth.AccessJwt
	xrpcc.Auth.RefreshJwt = auth.RefreshJwt
	xrpcc.Auth.Did = auth.Did
	xrpcc.Auth.Handle = auth.Handle

	return &client{client: xrpcc}, nil
}

func (c *client) Post(summary, telegramUrl string) error {
	const anchorText = "👉 Bekijk het volledige energiebericht op Telegram"

	text := fmt.Sprintf("%s\n\n%s", summary, anchorText)

	t := time.Now().UTC()
	ts := t.Format(time.RFC3339)
	rkey := string(syntax.NewTID(t.UnixMicro(), 42))

	if _, err := comatproto.RepoApplyWrites(context.Background(), c.client, &comatproto.RepoApplyWrites_Input{
		Repo: c.client.Auth.Did,
		Writes: []*comatproto.RepoApplyWrites_Input_Writes_Elem{
			{
				RepoApplyWrites_Create: &comatproto.RepoApplyWrites_Create{
					Collection: "app.bsky.feed.post",
					Rkey:       &rkey,
					Value: &lexutil.LexiconTypeDecoder{
						Val: &appbsky.FeedPost{
							CreatedAt: ts,
							Facets: []*appbsky.RichtextFacet{
								{
									Index: &appbsky.RichtextFacet_ByteSlice{
										ByteStart: int64(len(text) - len(anchorText)),
										ByteEnd:   int64(len(text)),
									},
									Features: []*appbsky.RichtextFacet_Features_Elem{
										{
											RichtextFacet_Link: &appbsky.RichtextFacet_Link{
												Uri: telegramUrl,
											},
										},
									},
								},
							},
							Langs: []string{"nl"},
							Tags:  []string{"energie", "energiebericht", "groen", "duurzaam", "stroom", "klimaat", "klimaatverandering"},
							Text:  text,
						},
					},
				},
			},
			{
				RepoApplyWrites_Create: &comatproto.RepoApplyWrites_Create{
					Collection: "app.bsky.feed.threadgate",
					Rkey:       &rkey,
					Value: &lexutil.LexiconTypeDecoder{
						Val: &appbsky.FeedThreadgate{
							Allow:     []*appbsky.FeedThreadgate_Allow_Elem{},
							CreatedAt: ts,
							Post:      fmt.Sprintf("at://%s/app.bsky.feed.post/%s", c.client.Auth.Did, rkey),
						},
					},
				},
			},
		},
	}); err != nil {
		return fmt.Errorf("bsky: apply writes: %w", err)
	}

	return nil
}
