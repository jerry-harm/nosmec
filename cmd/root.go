package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/logger"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
)

type appContextKey struct{}

var debug bool

var rootCmd = &cobra.Command{
	Use:   "nosmec",
	Short: "a cli for nostr",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			logger.SetDebug(true)
		}
	},
}

var app *config.AppContext

func Execute() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		if app != nil {
			app.Close()
		}
		os.Exit(0)
	}()

	err := rootCmd.Execute()
	if err != nil {
		handleError(err)
	}
}

func init() {
	cobra.OnInitialize(initApp)
	initCommands()

	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug file output")

	setupHTTPTransport()
}

func initApp() {
	cfg := config.InitConfig()
	config.SetProxyConfig(config.ProxyConfig{
		Socks:    cfg.Proxy.Socks,
		I2PSocks: cfg.Proxy.I2PSocks,
	})
	pool := config.GlobalPool()
	store := config.GlobalLMDB()
	app = config.NewAppContext(pool, store, cfg, config.GetViper())

	rootCmd.SetContext(context.WithValue(context.Background(), appContextKey{}, app))
}

func setupHTTPTransport() {
	transport := &http.Transport{
		Proxy: utils.ProxySelector,
	}
	http.DefaultTransport = transport
}

func getApp() *config.AppContext {
	return app
}

func getAppFromContext(ctx context.Context) *config.AppContext {
	if appPtr := ctx.Value(appContextKey{}); appPtr != nil {
		return appPtr.(*config.AppContext)
	}
	return app
}

func reloadApp() {
	if app != nil {
		app.Close()
	}
	cfg := config.InitConfig()
	config.SetProxyConfig(config.ProxyConfig{
		Socks:    cfg.Proxy.Socks,
		I2PSocks: cfg.Proxy.I2PSocks,
	})
	pool := config.GlobalPool()
	store := config.GlobalLMDB()
	app = config.NewAppContext(pool, store, cfg, config.GetViper())

	rootCmd.SetContext(context.WithValue(context.Background(), appContextKey{}, app))
}
