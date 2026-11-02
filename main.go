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
		ConnectionTimeout: 4 * time.Second, // set default to 4 seconds
		RequestTimeout: 4 * time.Second, // set default to 4 seconds
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

	//TODO: setup the context and the nessasary timeouts for client
	client := &http.Client{}
	ctx := context.TODO()

	mirrors, err := mirrorlist.Fetch(ctx, client)

	if err != nil{
		return err
	}

	mirrors = mirrors.Filter(appConfig.FilterOptions)

	if !appConfig.NoBenchmark{
		errs := mirrors.Benchmark(ctx, client, appConfig.Threads)

		if errs != nil{
			return errors.Join(errs...)
		}
	}

	mirrors, err = mirrors.Sort(appConfig.Sort)

	if err != nil{
		return err
	}

	mirrors, err = mirrors.Limit(appConfig.Number)

	if err != nil{
		return err
	}

	//TODO: handle the output write

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

