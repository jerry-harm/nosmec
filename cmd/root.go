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

type cmdContextKey struct{}

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
	// NOTE: GlobalPool() is NOT called here — pool/stores are lazily initialized
	// on first actual use (relay connection). This keeps shell completion fast
	// by avoiding LMDB open on every tab-press.
	app = config.NewAppContext(nil, cfg, config.GetViper())

	rootCmd.SetContext(context.WithValue(context.Background(), appContextKey{}, app))
	SetCmdApp(app)
}

func SetCmdApp(a *config.AppContext) {
	rootCmd.SetContext(context.WithValue(context.Background(), cmdContextKey{}, a))
}

func GetCmdApp() *config.AppContext {
	if appPtr := rootCmd.Context().Value(cmdContextKey{}); appPtr != nil {
		return appPtr.(*config.AppContext)
	}
	return app
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
	app = config.NewAppContext(pool, cfg, config.GetViper())

	rootCmd.SetContext(context.WithValue(context.Background(), appContextKey{}, app))
}
