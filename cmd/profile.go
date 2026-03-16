package cmd

import (
	"context"
	"fmt"
	"os"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile [npub or pubkey]",
	Short: "Query user profile information",
	Long: `Query Nostr profile information for a user.

If no argument is provided, query the current user's profile.
If an argument is provided, it can be an npub string or hex format pubkey.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var pubKey nostr.PubKey
		var err error

		// If no argument is provided, use the current user's public key
		if len(args) == 0 {
			pubKey, err = utils.GetMyPubKey()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to get current user's public key: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Parse user identifier (npub or hex pubkey)
			pubKey, err = parseUserIdentifier(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to parse user identifier: %v\n", err)
				os.Exit(1)
			}
		}

		// Query profile
		ctx := context.Background()
		event, err := utils.GetProfile(ctx, pubKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to query profile: %v\n", err)
			os.Exit(1)
		}

		// Output result (default to text format)
		utils.PrintEvent(event, false)
	},
}

// parseUserIdentifier parses user identifier, supports npub and hex pubkey
func parseUserIdentifier(identifier string) (nostr.PubKey, error) {
	// Try to decode as npub
	prefix, decoded, err := nip19.Decode(identifier)
	if err == nil {
		if prefix == "npub" {
			// npub decoded successfully, decoded should be nostr.PubKey type
			if pubKey, ok := decoded.(nostr.PubKey); ok {
				return pubKey, nil
			}
			// If type assertion fails, try to handle as string
			if hexStr, ok := decoded.(string); ok {
				var pubKey nostr.PubKey
				copy(pubKey[:], hexStr)
				return pubKey, nil
			}
			return nostr.PubKey{}, fmt.Errorf("failed to parse npub: unexpected type")
		} else {
			return nostr.PubKey{}, fmt.Errorf("unsupported bech32 prefix: %s, expected npub", prefix)
		}
	}

	// nip19.Decode failed, try to handle as hex pubkey
	var pubKey nostr.PubKey
	if len(identifier) == 64 { // hex pubkey is typically 64 characters
		// Validate if it's a valid hex string
		for _, c := range identifier {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return nostr.PubKey{}, fmt.Errorf("invalid hex pubkey: contains non-hex characters")
			}
		}
		copy(pubKey[:], identifier)
		return pubKey, nil
	}

	return nostr.PubKey{}, fmt.Errorf("invalid user identifier, expected npub or 64-character hex pubkey")
}
