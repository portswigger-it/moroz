package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	//"github.com/go-kit/kit/log/level"

	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/kolide/kit/env"
	"github.com/kolide/kit/httputil"
	"github.com/kolide/kit/version"
	"github.com/oklog/run"

	"github.com/groob/moroz/logging"
	"github.com/groob/moroz/moroz"
	"github.com/groob/moroz/santaconfig"
)

const openSSLBash = `
Looks like you're missing a TLS certificate and private key. You can quickly generate one 
by using the commands below:

	./tools/dev/certificate/create

Add the santa hostname to your hosts file.

	sudo echo "127.0.0.1 santa" >> /etc/hosts

And then, add the cert to roots.

	./tools/dev/certificate/add-trusted-cert


The latest version of santa is available on the github repo page:
	https://github.com/google/santa/releases
`

func main() {
	var (
		flTLSCert       = flag.String("tls-cert", env.String("MOROZ_TLS_CERT", "server.crt"), "path to TLS certificate")
		flTLSKey        = flag.String("tls-key", env.String("MOROZ_TLS_KEY", "server.key"), "path to TLS private key")
		flAddr          = flag.String("http-addr", env.String("MOROZ_HTTP_ADDRESS", ":8080"), "http address ex: -http-addr=:8080")
		flConfigs       = flag.String("configs", env.String("MOROZ_CONFIGS", "../../configs"), "path to config folder")
		flEvents        = flag.String("event-dir", env.String("MOROZ_EVENT_DIR", "/tmp/santa_events"), "Path to root directory where events will be stored.")
		flPersistEvents = flag.Bool("persist-events", env.Bool("MOROZ_WRITE_EVENTS", true), "Enable or disable event persistence to disk. Defaults to enabled.")
		flVersion       = flag.Bool("version", false, "print version information")
		flDebug         = flag.Bool("debug", false, "log at a debug level by default.")
		flUseTLS        = flag.Bool("use-tls", true, "I promise I terminated TLS elsewhere when changing this")
	)
	flag.Parse()

	if *flVersion {
		version.PrintFull()
		return
	}

	// Initialize the logger
	logging.InitLogger(*flDebug)
	logging.Logger.Log("msg", "Application started")

	if _, err := os.Stat(*flTLSCert); *flUseTLS && os.IsNotExist(err) {
		logging.Logger.Log("level", "info", "msg", openSSLBash)
		logging.Logger.Log("level", "info", "msg", "you need to provide at least a 'global.toml' configuration file in the configs folder. See the configs folder in the git repo for an example")
		os.Exit(2)
	}

	if !validateConfigExists(*flConfigs) {
		logging.Logger.Log("level", "error", "msg", "you need to provide at least a 'global.toml' configuration file in the configs folder. See the configs folder in the git repo for an example")
		os.Exit(2) // Exit with a specific status code
	}

	repo := santaconfig.NewFileRepo(*flConfigs)
	var svc moroz.Service
	{
		s, err := moroz.NewService(repo, *flEvents, *flPersistEvents, logging.Logger)
		if err != nil {
			logging.Logger.Log("level", "error", "msg", "Failed to create service", "err", err)
			os.Exit(1)
		}
		svc = s
		svc = moroz.LoggingMiddleware(logging.Logger)(svc)
	}

	endpoints := moroz.MakeServerEndpoints(svc)

	r := mux.NewRouter()
	moroz.AddHTTPRoutes(r, endpoints, logging.Logger)

	var g run.Group
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				err := fmt.Errorf("received signal %s", sig)
				logging.Logger.Log("level", "error", "msg", "Received signal", "signal", sig, "err", err)
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		}, func(error) {
			cancel()
		})
	}

	{
		srv := httputil.NewServer(*flAddr, r)
		g.Add(func() error {
			//level.Debug(logger).Log("msg", "serve http", "tls", *flUseTLS, "addr", *flAddr)
			level.Debug(logging.Logger).Log("msg", "serve http", "tls", *flUseTLS, "addr", *flAddr)
			if *flUseTLS {
				return srv.ListenAndServeTLS(*flTLSCert, *flTLSKey)
			} else {
				return srv.ListenAndServe()
			}
		}, func(error) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			srv.Shutdown(ctx)
		})
	}

	//logutil.Fatal(logger, "msg", "terminated", "err", g.Run())
	err := g.Run()
	if err != nil {
		logging.Logger.Log("level", "error", "msg", "terminated", "err", err)
		os.Exit(1) // Exit with a non-zero status to indicate failure
	}
}

func validateConfigExists(configsPath string) bool {
	var hasConfig = true
	if _, err := os.Stat(configsPath); os.IsNotExist(err) {
		hasConfig = false
	}
	if _, err := os.Stat(configsPath + "/global.toml"); os.IsNotExist(err) {
		hasConfig = false
	}
	if !hasConfig {
	}
	return hasConfig
}
