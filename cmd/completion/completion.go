package completion

import (
	"strings"

	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
)

type cmdContextKey struct{}

func GetApp(cmd *cobra.Command) *config.AppContext {
	if appPtr := cmd.Root().Context().Value(cmdContextKey{}); appPtr != nil {
		return appPtr.(*config.AppContext)
	}
	return nil
}

func AliasCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	app := GetApp(cmd)
	if app == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	aliases := utils.ListAliases(app)
	var completions []string
	for alias := range aliases {
		if strings.HasPrefix(alias, toComplete) {
			completions = append(completions, alias)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func PubKeyCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	app := GetApp(cmd)
	if app == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string

	aliases := utils.ListAliases(app)
	for alias := range aliases {
		if strings.HasPrefix(alias, toComplete) {
			completions = append(completions, alias)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func RelayCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	app := GetApp(cmd)
	if app == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	relays := app.ListRelays()
	for _, r := range relays {
		if strings.HasPrefix(r.URL, toComplete) {
			completions = append(completions, r.URL)
		}
	}

	if len(completions) == 0 {
		for _, r := range relays {
			completions = append(completions, r.URL)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func CommunityCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	app := GetApp(cmd)
	if app == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string

	communities := app.ListSubscriptions("community")
	for _, c := range communities {
		if strings.HasPrefix(c.ID, toComplete) {
			completions = append(completions, c.ID)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func HashtagCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	app := GetApp(cmd)
	if app == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string

	hashtags := app.ListSubscriptions("hashtag")
	for _, h := range hashtags {
		hashtag := h.ID
		if !strings.HasPrefix(hashtag, "#") {
			hashtag = "#" + hashtag
		}
		if strings.HasPrefix(hashtag, toComplete) {
			completions = append(completions, hashtag)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func BotFlagCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp
}

func LimitCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"10", "20", "50", "100"}, cobra.ShellCompDirectiveNoFileComp
}

func FilterCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"followed", "global", "mine"}, cobra.ShellCompDirectiveNoFileComp
}

func SubTypeCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	app := GetApp(cmd)
	if app == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var types []string
	if len(app.ListSubscriptions("user")) > 0 {
		types = append(types, "user")
	}
	if len(app.ListSubscriptions("community")) > 0 {
		types = append(types, "community")
	}
	if len(app.ListSubscriptions("hashtag")) > 0 {
		types = append(types, "hashtag")
	}

	var completions []string
	for _, t := range types {
		if strings.HasPrefix(t, toComplete) {
			completions = append(completions, t)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func ConfigKeyCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	keys := []string{
		"private_key",
		"data_dir",
		"proxy.socks",
		"proxy.i2p_socks",
		"relay_list",
		"dm_relays",
		"search_relays",
		"known_relays",
		"alias",
	}

	var completions []string
	for _, k := range keys {
		if strings.HasPrefix(k, toComplete) {
			completions = append(completions, k)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func ConfigRelayURLCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	app := GetApp(cmd)
	if app == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	relays := app.ListRelays()
	for _, r := range relays {
		if strings.HasPrefix(r.URL, toComplete) {
			completions = append(completions, r.URL)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func GlobalCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp
}
