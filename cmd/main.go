package main

import (
	"flag"
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc/shared"
	"github.com/spf13/viper"
	"github.com/lomomike/jaeger-postgresql/pgstore"
)

var configPath string

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "jaeger-postgresql",
		Level: hclog.Warn, // Jaeger only captures >= Warn, so don't bother logging below Warn
	})

	flag.StringVar(&configPath, "config", "", "The absolute path to the Postgresql plugin's configuration file")
	flag.Parse()

	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	if configPath != "" {
		v.SetConfigFile(configPath)

		err := v.ReadInConfig()
		if err != nil {
			logger.Error("failed to parse configuration file", "error", err)
			os.Exit(1)
		}
	}

	conf := &Configuration{}
	conf.InitFromViper(v)

	var store shared.StoragePlugin
	var closeStore func() error
	var err error

	store, closeStore, err = NewStore(conf, logger)

	if err != nil {
		logger.Error("failed to open store", "error", err)
		os.Exit(1)
	}

	grpc.Serve(store)

	if err = closeStore(); err != nil {
		logger.Error("failed to close store", "error", err)
		os.Exit(1)
	}
}
