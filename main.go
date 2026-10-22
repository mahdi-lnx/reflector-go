
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"github.com/mahdi-lnx/reflector-go/internals/cli"
)

func main(){
	appConfig := cli.Config{
		ConnectionTimeout: 4,
		DownloadTimeout: 4,
		Threads: runtime.NumCPU(),
		Sort: cli.DownloadRate,
		Format: cli.Mirrorlist,
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

