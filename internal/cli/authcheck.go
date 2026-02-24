package cli

import (
	"errors"

	"oe-cli/internal/appleads"
)

const missingCredsMessage = "Missing Apple Ads credentials in env. Set OE_ADS_CREDENTIALS_JSON or OE_ADS_CLIENT_ID/OE_ADS_TEAM_ID/OE_ADS_KEY_ID/OE_ADS_PRIVATE_KEY"

func ensureCredentialsPresent() error {
	creds, err := appleads.LoadCredentials()
	if err != nil {
		return err
	}
	if creds == nil || !creds.IsComplete() {
		return errors.New(missingCredsMessage)
	}
	return nil
}
