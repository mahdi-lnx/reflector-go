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

func run() error{

	appConfig := cli.Config{
		RequestTimeout: 7 * time.Second, // set default to 7 seconds
		Threads: 47,
		Sort: mirrorlist.DownloadRate,
		Format: mirrorlist.PacmanFormat,
		Protocol: "https",
		MinCompletion: 100,
		ActiveOnly: true,
	}

	err := appConfig.Parse(os.Args[1:])

	if err != nil{
		return err
	}

	client := &http.Client{
		Timeout: appConfig.RequestTimeout,
	}

	ctx := context.TODO()

	logHandlerOptions := &slog.HandlerOptions{ Level: slog.LevelError, }

	if appConfig.Verbose{
		logHandlerOptions.Level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, logHandlerOptions))


	logger.Info("Fetching mirrors")

	totalFetchedMirrors, err := mirrorlist.Fetch(ctx, client)

	if err != nil{
		return err
	}

	logger.Info("Fetched mirrors", "count", len(totalFetchedMirrors))

	logger.Info("Filtering mirrors", "count", len(totalFetchedMirrors))

	filteredMirrors := totalFetchedMirrors.Filter(appConfig.FilterOptions)

	if len(filteredMirrors) == 0{
		// no mirror passed the filter

		return errors.New("no mirrors matched the specified filters")
	}

	logger.Info(
		"Filtered mirrors", 
		"matched", len(filteredMirrors), 
		"excluded", len(totalFetchedMirrors) - len(filteredMirrors),
	)

	benchmarkedMirrors := filteredMirrors

	if !appConfig.NoBenchmark{
		//TODO: need to decide how to handle the errs if appConfig.verbosity is true
		
		// t := time.Now() TODO: need to remove this

		logger.Info("Benchmarking mirrors", "count", len(filteredMirrors), "workers", appConfig.Threads)

		result, errs := filteredMirrors.Benchmark(ctx, client, appConfig.Threads)

		benchmarkedMirrors = result

		// fmt.Println(time.Since(t).Seconds()) TODO: need to remove this

		if err := ctx.Err(); err != nil{
			return fmt.Errorf("reflector was intrupted: %w", err)
		}

		if len(benchmarkedMirrors) == 0{
			return errors.New("no successfully benchmarked mirrors")
		}

		for _, err := range errs {
			logger.Warn("Mirror Benchmark failed", "error", err.Error())
		}

		logger.Info(
			"Benchmarked mirrors", 
			"success", len(benchmarkedMirrors), 
			"failed", len(errs),
		)
	}

	logger.Info("Sorting mirrors", "by", appConfig.Sort)

	sortedMirrors, err := benchmarkedMirrors.Sort(appConfig.Sort)

	if err != nil{
		return err
	}

	logger.Info("Sorted mirrors", "count", len(sortedMirrors))

	finalMirrors, err := sortedMirrors.Limit(appConfig.Number)

	if err != nil{
		return err
	}

	logger.Info("Selected mirrors", "count", len(finalMirrors))

	if appConfig.Save == ""{
		// write the mirrorlist to stdout
		err := finalMirrors.Write(os.Stdout, appConfig.Format)

		if err != nil{
			return err
		}
		return nil
	}

	// on a fresh installed arch system the premission mode
	// of /etc/pacman.d/mirrorlist is 0o644
	f, err := atomicfile.NewAtomicFile(appConfig.Save, 0o644)

	if err != nil{
		return err
	}

	// ignore the error returned by Close for now
	defer f.Close()

	err = finalMirrors.Write(f, appConfig.Format)

	if err != nil{
		return err
	}

	err = f.Commit()

	if err != nil{
		return err
	}

	logger.Info("Saved mirrorlist", "path", appConfig.Save, "count", len(finalMirrors))

	return nil
}


func main(){
	err := run()

	if err != nil{
		if errors.Is(err, flag.ErrHelp){
			fmt.Print(err)
			return
		}

		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

