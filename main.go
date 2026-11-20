package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mahdi-lnx/reflector-go/internals/cli"
	"github.com/mahdi-lnx/reflector-go/internals/mirrorlist"
)

func main() {

	startTime := time.Now()

	appConfig := cli.Config{
		RequestTimeout: 7 * time.Second, // set default to 7 seconds
		Threads:        47,
		Sort:           mirrorlist.DownloadRate,
		Format:         mirrorlist.PacmanFormat,
		Protocol:       "https",
		MinCompletion:  100,
		ActiveOnly:     true,
	}

	err := appConfig.Parse(os.Args[1:])

	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Print(err)
			return
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: appConfig.RequestTimeout,
	}

	ctx := context.TODO()

	logger := newAppLogger(appConfig.Verbose)

	err = run(ctx, client, appConfig, logger)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	logger.Info(
		"Mirrorlist generation completed",
		"duration", time.Since(startTime),
	)
}
