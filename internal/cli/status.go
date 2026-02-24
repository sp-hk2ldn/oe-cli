package cli

import (
	"context"
	"fmt"
	"time"

	"oe-cli/internal/appleads"
)

func RunStatus(ctx context.Context) {
	now := time.Now().UTC().Format(time.RFC3339)
	creds, err := appleads.LoadCredentials()
	if err != nil {
		fmt.Println("oe-ads status")
		fmt.Printf("time=%s\n", now)
		fmt.Println("credentials=invalid")
		fmt.Printf("error=%s\n", err.Error())
		return
	}

	if creds == nil || !creds.IsComplete() {
		fmt.Println("oe-ads status")
		fmt.Printf("time=%s\n", now)
		fmt.Println("credentials=missing")
		fmt.Println("next=set OE_ADS_CREDENTIALS_JSON or OE_ADS_CLIENT_ID/OE_ADS_TEAM_ID/OE_ADS_KEY_ID/OE_ADS_PRIVATE_KEY")
		return
	}

	client := appleads.NewClient(nil)
	orgID, err := client.ValidateCredentials(ctx)
	if err != nil {
		fmt.Println("oe-ads status")
		fmt.Printf("time=%s\n", now)
		fmt.Println("credentials=present")
		fmt.Println("auth=failed")
		fmt.Printf("error=%s\n", err.Error())
		return
	}

	fmt.Println("oe-ads status")
	fmt.Printf("time=%s\n", now)
	fmt.Println("credentials=ok")
	fmt.Printf("orgId=%s\n", orgID)
	if creds.OrgID != "" {
		fmt.Printf("configuredOrgId=%s\n", creds.OrgID)
	}
}
