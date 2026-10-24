
// Package cli provides the necessary cli apis for this app
package cli

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/mahdi-lnx/reflector-go/internals/mirrorlist"
)


// A Config specifies the config flags passed to the app
//
// The values set to each flag on init specify the default
// value of each flag
type Config struct{
	Verbose bool // increases logging verbosity
	ConnectionTimeout time.Duration // the duration to wait before a connection times out
	DownloadTimeout time.Duration // the duration to wait before a download times out
	Save string // a file path to save the mirrorlist into
	Threads int // The number of threads used for rating the download speed of the mirrors

	// specifes a sort cirteria to sort the Mirrolist by
	//
	// Possible values:
	// 	- [mirrorlist.DownloadRate]
	// 	- [mirrorlist.Age]
	// 	- [mirrorlist.Score]
	Sort mirrorlist.SortCriteria  

	// specifies the output format
	//
	// Possible values:
	//	- [mirrorlist.MirrorlistFormat]
	//	- [mirrorlist.JsonFormat]
	//	- [mirrorlist.TableFormat]
	Format mirrorlist.FormatOption

	mirrorlist.FilterOptions
}


// A Sentinel error that indicates invalid configuration
var ErrInvalidConfig error = errors.New("invalid configuration")


// ConfigError represents an error
// that was encountered when parsing
// the args given to [Config.Parse]
type ConfigError struct{
	strings.Builder
	cause error
}

// Error Returns the a formated string error
func (ce *ConfigError) Error() string{
	return ce.String()
}

// Unwrap returns the cause of the ConfigError
// resulting in the original error to be preserved
func (ce *ConfigError) Unwrap() error{
	return ce.cause
}

// LoadConfig loads the flags passed to the app
// with the specified defaults being the default values
// for each flag
//
// Note: if the returned error is not nil, the values
// of c must not be reliad on
func (c *Config) Parse(args []string) (error){
	flagSet := flag.NewFlagSet("reflector", flag.ContinueOnError)

	// set the verbosity flag
	flagSet.BoolVar(&c.Verbose, "verbose", c.Verbose, "Increases the logging verbosity")

	// set timeout flags
	flagSet.DurationVar(&c.ConnectionTimeout, "connection-timeout", c.ConnectionTimeout, "The duration to wait before a connection times out")
	flagSet.DurationVar(&c.DownloadTimeout, "download-timeout", c.DownloadTimeout, "The duration to wait before a download times out")

	// set the file path flag
	flagSet.StringVar(&c.Save, "save", c.Save, "The file path to save the output into")

	// set the format flag
	var mirrorlistFormat string
	flagSet.StringVar(&mirrorlistFormat, "format", string(c.Format), "The output format")

	// set the sort flag
	var sortCriteria string
	flagSet.StringVar(&sortCriteria, "sort", string(c.Sort), "The cirteria to sort the Mirrolist by")

	// set the threads flag
	flagSet.IntVar(&c.Threads, "threads", c.Threads, "The number of threads used for rating the download speed of the mirrors")

	// set the filtering flags
	flagSet.StringVar(&c.Country, "country", c.Country, "Only return mirrors belonging to the specified country")
	flagSet.DurationVar(&c.Age, "age", c.Age, "Only return mirrors that have synced at least in the last specified hours")
	flagSet.IntVar(&c.Number, "number", c.Number, "Only return at most n mirrors")

	accErr := &ConfigError{}

	flagSet.SetOutput(accErr)

	err := flagSet.Parse(args)

	// update c.Sort and c.Format
	c.Sort = mirrorlist.SortCriteria(sortCriteria)
	c.Format = mirrorlist.FormatOption(mirrorlistFormat)

	if err != nil {
		accErr.cause = err
		return accErr
	}

	c.validate(accErr)

	if accErr.Len() > 0{
		accErr.cause = ErrInvalidConfig
		return accErr
	}

	return nil
}

// validate validates the [Config]
func (c *Config) validate(configError *ConfigError){

	if c.Threads <= 0{
		configError.WriteString("Invalid value for option 'threads': value must be greater than 0\n")
	}

	if c.DownloadTimeout.Seconds() <= 0{
		configError.WriteString("Invalid value for option 'download-timeout': value must be greater than 0\n")
	}

	if c.ConnectionTimeout.Seconds() <= 0{
		configError.WriteString("Invalid value for option 'connection-timeout': value must be greater than 0\n")
	}

	if c.Age < 0{
		configError.WriteString("Invalid value for option 'age': value must be greater than or equal to 0\n")
	}

	if c.Number < 0{
		configError.WriteString("Invalid value for option 'number': value must be greater than to or equal to 0\n")
	}

	if c.Format != mirrorlist.MirrorlistFormat && c.Format != mirrorlist.JsonFormat && c.Format != mirrorlist.TableFormat{
		fmt.Fprintf(
			configError,
			"Invalid value for option 'format': value must be one of '%s', '%s', or '%s'\n",
			mirrorlist.TableFormat,
			mirrorlist.MirrorlistFormat,
			mirrorlist.JsonFormat,
		)
	}

	if c.Sort != mirrorlist.DownloadRate && c.Sort != mirrorlist.Score && c.Sort != mirrorlist.Age {
			fmt.Fprintf(
				configError, 
				"Invalid value for option 'sort': value must be one of '%s', '%s', or '%s'\n", 
				mirrorlist.Score,
				mirrorlist.Age,
				mirrorlist.DownloadRate,
			)
	}
}

