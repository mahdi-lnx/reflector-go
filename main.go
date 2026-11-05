package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
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

	mirrors, err := mirrorlist.Fetch(ctx, client)

	if err != nil{
		return err
	}

	mirrors = mirrors.Filter(appConfig.FilterOptions)

	if len(mirrors) == 0{
		// no mirror passed the filter

		return errors.New("no mirrors matched the specified filters")
	}

	if !appConfig.NoBenchmark{
		//TODO: need to decide how to handle the errs if appConfig.verbosity is true
		
		// t := time.Now() TODO: need to remove this

		benchmarkedMirrors, _ := mirrors.Benchmark(ctx, client, appConfig.Threads)

		// fmt.Println(time.Since(t).Seconds()) TODO: need to remove this

		if err := ctx.Err(); err != nil{
			return fmt.Errorf("reflector was intrupted: %w", err)
		}

		if len(benchmarkedMirrors) == 0{
			return errors.New("no successfully benchmarked mirrors")
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

	// on a fresh installed arch system the premission mode
	// of /etc/pacman.d/mirrorlist is 0o644
	f, err := atomicfile.NewAtomicFile(appConfig.Save, 0o644)

	if err != nil{
		return err
	}

	// ignore the error returned by Close for now
	defer f.Close()

	err = mirrors.Write(f, appConfig.Format)

	if err != nil{
		return err
	}

	return f.Commit()
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

