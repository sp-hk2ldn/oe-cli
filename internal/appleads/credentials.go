package appleads

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const credentialsEnvJSON = "OE_ADS_CREDENTIALS_JSON"

type Credentials struct {
	ClientID          string `json:"clientId"`
	TeamID            string `json:"teamId"`
	KeyID             string `json:"keyId"`
	PrivateKey        string `json:"privateKey"`
	OrgID             string `json:"orgId,omitempty"`
	PopularityAdamID  string `json:"popularityAdamId,omitempty"`
	PopularityAdGroup int    `json:"popularityAdGroupId,omitempty"`
	PopularityCookie  string `json:"popularityWebCookie,omitempty"`
	PopularityXSRF    string `json:"popularityXsrfToken,omitempty"`
}

func (c Credentials) IsComplete() bool {
	return strings.TrimSpace(c.ClientID) != "" &&
		strings.TrimSpace(c.TeamID) != "" &&
		strings.TrimSpace(c.KeyID) != "" &&
		strings.TrimSpace(c.PrivateKey) != ""
}

func LoadCredentials() (*Credentials, error) {
	if raw := strings.TrimSpace(os.Getenv(credentialsEnvJSON)); raw != "" {
		var creds Credentials
		if err := json.Unmarshal([]byte(raw), &creds); err != nil {
			return nil, fmt.Errorf("invalid %s JSON: %w", credentialsEnvJSON, err)
		}
		if !creds.IsComplete() {
			return nil, errors.New("incomplete credentials in OE_ADS_CREDENTIALS_JSON")
		}
		return &creds, nil
	}

	clientID := strings.TrimSpace(os.Getenv("OE_ADS_CLIENT_ID"))
	teamID := strings.TrimSpace(os.Getenv("OE_ADS_TEAM_ID"))
	keyID := strings.TrimSpace(os.Getenv("OE_ADS_KEY_ID"))
	privateKey := strings.TrimSpace(os.Getenv("OE_ADS_PRIVATE_KEY"))

	if clientID == "" || teamID == "" || keyID == "" || privateKey == "" {
		return nil, nil
	}

	adGroupID, _ := strconv.Atoi(strings.TrimSpace(os.Getenv("OE_ADS_POPULARITY_ADGROUP_ID")))
	creds := &Credentials{
		ClientID:          clientID,
		TeamID:            teamID,
		KeyID:             keyID,
		PrivateKey:        privateKey,
		OrgID:             strings.TrimSpace(os.Getenv("OE_ADS_ORG_ID")),
		PopularityAdamID:  strings.TrimSpace(os.Getenv("OE_ADS_POPULARITY_ADAM_ID")),
		PopularityAdGroup: adGroupID,
		PopularityCookie:  strings.TrimSpace(os.Getenv("OE_ADS_POPULARITY_WEB_COOKIE")),
		PopularityXSRF:    strings.TrimSpace(os.Getenv("OE_ADS_POPULARITY_XSRF_TOKEN")),
	}
	return creds, nil
}
