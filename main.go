package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/mahdi-lnx/reflector-go/internals/atomicfile"
	"github.com/mahdi-lnx/reflector-go/internals/cli"
	"github.com/mahdi-lnx/reflector-go/internals/mirrorlist"
)

// writeOutput writes mirrors to stdout or atomically saves them to a file
func writeOutput(mirrors mirrorlist.Mirrorlist, config cli.Config, logger *slog.Logger) error {
	if config.Save == "" {
		// write the mirrorlist to stdout
		return mirrors.Write(os.Stdout, config.Format)
	}

	// on a fresh installed arch system the premission mode
	// of /etc/pacman.d/mirrorlist is 0o644
	f, err := atomicfile.NewAtomicFile(config.Save, 0o644)

	if err != nil {
		return err
	}

	// ignore the error returned by Close for now
	defer f.Close()

	err = mirrors.Write(f, config.Format)

	if err != nil {
		return err
	}

	err = f.Commit()

	if err != nil {
		return err
	}

	logger.Info("Saved mirrorlist", "path", config.Save, "count", len(mirrors))

	return nil
}

// benchmarkMirrors benchmarks the given mirrors unless benchmarking is disabled
//
// It logs individual benchmark failures and returns the successfully benchmarked
// mirrors. If benchmarking is disabled, it returns the original mirrors unchanged
//
// It returns an error if benchmarking is interrupted via ctx or no mirrors succeed
func benchmarkMirrors(
	ctx context.Context,
	client *http.Client,
	mirrors mirrorlist.Mirrorlist,
	config cli.Config,
	logger *slog.Logger,
) (mirrorlist.Mirrorlist, error) {
	if !config.NoBenchmark {

		logger.Info("Benchmarking mirrors", "count", len(mirrors), "workers", config.Threads)

		benchmarkedMirrors, errs := mirrors.Benchmark(ctx, client, config.Threads)

		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("reflector was intrupted: %w", err)
		}

		for _, err := range errs {
			logger.Warn("Mirror benchmark failed", "error", err)
		}

		if len(benchmarkedMirrors) == 0 {
			return nil, errors.New("no successfully benchmarked mirrors")
		}

		logger.Info(
			"Benchmarked mirrors",
			"success", len(benchmarkedMirrors),
			"failed", len(errs),
		)

		return benchmarkedMirrors, nil
	}

	return mirrors, nil
}

func run(ctx context.Context, client *http.Client, config cli.Config, logger *slog.Logger) error {

	logger.Info("Fetching mirrors")

	totalFetchedMirrors, err := mirrorlist.Fetch(ctx, client)

	if err != nil {
		return err
	}

	logger.Info("Fetched mirrors", "count", len(totalFetchedMirrors))

	logger.Info("Filtering mirrors", "count", len(totalFetchedMirrors))

	filteredMirrors := totalFetchedMirrors.Filter(config.FilterOptions)

	if len(filteredMirrors) == 0 {
		// no mirror passed the filter

		return errors.New("no mirrors matched the specified filters")
	}

	logger.Info(
		"Filtered mirrors",
		"matched", len(filteredMirrors),
		"excluded", len(totalFetchedMirrors)-len(filteredMirrors),
	)

	processedMirrors, err := benchmarkMirrors(ctx, client, filteredMirrors, config, logger)

	if err != nil {
		return err
	}

	logger.Info("Sorting mirrors", "by", config.Sort)

	sortedMirrors, err := processedMirrors.Sort(config.Sort)

	if err != nil {
		return err
	}

	finalMirrors, err := sortedMirrors.Limit(config.Number)

	if err != nil {
		return err
	}

	logger.Info("Selected mirrors", "count", len(finalMirrors))

	return writeOutput(finalMirrors, config, logger)
}

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
