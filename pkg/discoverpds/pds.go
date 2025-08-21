// Package discoverpds finds the Personal Data Server for a given AT Protocol handle.
package discoverpds

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/xrpc"
)

var errNoPDS = errors.New("no PDS found in DID document")

// PDS finds the Personal Data Server for the given AT Protocol handle.
func PDS(ctx context.Context, handle string) (string, error) {
	client := &xrpc.Client{
		Host: "https://public.api.bsky.app",
	}

	rho, err := atproto.IdentityResolveHandle(ctx, client, handle)
	if err != nil {
		return "", fmt.Errorf("invoke com.atproto.identity.resolveHandle: %w", err)
	}

	resp, err := http.Get(fmt.Sprintf("https://plc.directory/%s", rho.Did))
	if err != nil {
		return "", fmt.Errorf("resolve DID %q: %w", rho.Did, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var doc identity.DIDDocument

	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return "", fmt.Errorf("decode DID document: %w", err)
	}

	if rho.Did != doc.DID.String() {
		return "", fmt.Errorf("DID mismatch: %q != %q", rho.Did, doc.DID.String())
	}

	pds, err := pds(doc.Service)
	if err != nil {
		return "", fmt.Errorf("find PDS in DID document: %w", err)
	}

	return pds, nil
}

func pds(services []identity.DocService) (string, error) {
	for _, service := range services {
		if service.Type == "AtprotoPersonalDataServer" {
			return service.ServiceEndpoint, nil
		}
	}

	return "", errNoPDS
}
