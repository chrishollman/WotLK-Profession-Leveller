package tsm

import (
	"bytes"
	"encoding/json"
	"fmt"

	rhttp "github.com/hashicorp/go-retryablehttp"
)

func (ts *TSMService) auth() {
	// Auth payload
	payload := &tsmAuthReq{
		ClientID:  authClientID,
		GrantType: authGrantType,
		Scope:     authScope,
		Token:     ts.cfg.apiKey,
	}
	payloadBytes, _ := json.Marshal(payload)

	// Make request
	req, err := rhttp.NewRequest("POST", tsmAuthURL, bytes.NewReader(payloadBytes))
	if err != nil {
		ts.logger.Fatalf("couldn't make an auth request to TSM: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	res, err := ts.client.Do(req)
	if err != nil {
		ts.logger.Fatalf("couldn't send request to TSM API: %v", err)
	}

	// Parse response
	resPayload := &tsmAuthRes{}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(resPayload)
	if err != nil {
		// TODO: Something better
		panic("can't handle TSM Auth response")
	}

	ts.cfg.authToken = resPayload.AccessToken
	ts.cfg.expiry = resPayload.Expiry
}

func (ts *TSMService) getBearer() string {
	if ts.cfg.authToken == "" {
		ts.logger.Error("Getting an empty bearer token")
	}

	return fmt.Sprintf("Bearer %v", ts.cfg.authToken)
}
