
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/mahdi-lnx/reflector-go/internals/cli"
	"github.com/mahdi-lnx/reflector-go/internals/mirrorlist"
)

func main(){
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
			fmt.Print(err)
			return
		}

		fmt.Println(err)

		os.Exit(1)
	}
}

