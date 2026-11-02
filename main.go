package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
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
		if errors.Is(err, flag.ErrHelp){
			return errors.New(strings.TrimRight(err.Error(), "\n"))
		}

		return err
	}



	return nil
}

func main(){
	err := run()

	if err != nil{
		fmt.Println(err)
		os.Exit(1)
	}
}

