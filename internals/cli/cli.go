
// Package cli provides the necessary cli apis for this app
package cli

import (
	"errors"
	"flag"
	"strings"
	"time"
)

// MirrorlistFilterConfig specifies the filters applied to a Mirrorlist
type MirrorlistFilterConfig struct{
	// a min number, in hours, that a mirror's last sync time must be less than or equal to
	//
	// Note: 0 indicates any sync time is fine
	Age float64 

	Country string // a country that a mirror must belong to

	// the number of enteries in the Mirrorlist
	//
	// Note: 0 indicates unlimited enteries
	Number int 
}


const (
	Mirrorlist = "mirrorlist" 
	Json = "json" 
	Table = "table" 
)


const (
	Age = "age" 
	DownloadRate = "download-rate" 
	Score = "score" 
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
	// 	- [DownloadRate]
	// 	- [Age]
	// 	- [Score]
	Sort string  

	// specifies the output format
	//
	// Possible values:
	//	- [Mirrorlist]
	//	- [Json]
	//	- [Table]
	Format string

	MirrorlistFilterConfig
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

func (ce *ConfigError) Unwrap() error{
	return ce.cause
}

// LoadConfig loads the flags passed to the app
// with the specified defaults being the default values
// for each flag
//
// LoadConfig internally uses [flag.Parse] under the hood
// to parse the flags; hence, if the parsed flags are incorrect,
// the error is handled the usual way [flag.Parse] handles it
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
	flagSet.StringVar(&c.Format, "format", c.Format, "The output format")

	// set the sort flag
	flagSet.StringVar(&c.Sort, "sort", c.Sort, "The cirteria to sort the Mirrolist by")

	// set the threads flag
	flagSet.IntVar(&c.Threads, "threads", c.Threads, "The number of threads used for rating the download speed of the mirrors")

	// set the filtering flags
	flagSet.StringVar(&c.Country, "country", c.Country, "Only return mirrors belonging to the specified country")
	flagSet.Float64Var(&c.Age, "age", c.Age, "Only return mirrors that have synced at least in the last specified hours")
	flagSet.IntVar(&c.Number, "number", c.Number, "Only return at most n mirrors")

	accErr := &ConfigError{}

	flagSet.SetOutput(accErr)

	err := flagSet.Parse(args)

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

	if c.Format != "table" && c.Format != "mirrorlist" && c.Format != "json"{
		configError.WriteString("Invalid value for option 'format': value must be one of 'table', 'mirrorlist', or 'json'\n")
	}

	if c.Sort != "score" && c.Sort != "age" && c.Sort != "download-rate"{
			configError.WriteString("Invalid value for option 'sort': value must be one of 'score', 'age', or 'download-rate'\n")
	}
}

