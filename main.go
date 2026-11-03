package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/mahdi-lnx/reflector-go/internals/cli"
	"github.com/mahdi-lnx/reflector-go/internals/mirrorlist"
)

func run() error{
	//TODO: implement this function

	appConfig := cli.Config{
		RequestTimeout: 4 * time.Second, // set default to 7 seconds
		Threads: runtime.NumCPU(),
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

	mirrors, err := mirrorlist.Fetch(ctx, client)

	if err != nil{
		return err
	}

	mirrors = mirrors.Filter(appConfig.FilterOptions)

	if !appConfig.NoBenchmark{
		benchmarkedMirrors, errs := mirrors.Benchmark(ctx, client, appConfig.Threads)

		//TODO: handle len(benchmarkedMirrors) == 0

		if errs != nil{
			if ctx.Err() != nil{
				// TODO: gracefull shutdown
			}

			//TODO: handle other errs

		}

		mirrors = benchmarkedMirrors
	}

	// its impossible for the following to return non-nil errors
	mirrors, _ = mirrors.Sort(appConfig.Sort)
	mirrors, _ = mirrors.Limit(appConfig.Number)

	if appConfig.Save == ""{
		// write the mirrorlist to stdout
		err := mirrors.Write(os.Stdout, appConfig.Format)

		if err != nil{
			return err
		}
		return nil
	}

	// on a fresh installed arch system /etc/pacman.d/mirrorlist has a mode of 0o644
	file, err := os.OpenFile(appConfig.Save, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0o644)

	if err != nil{
		return err
	}

	err = mirrors.Write(file, appConfig.Format)

	if err != nil{
		return err
	}

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

