// Package cli provides the necessary cli apis for this app
package cli

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/mahdi-lnx/reflector-go/internals/mirrorlist"
	"github.com/mahdi-lnx/reflector-go/internals/multivalueflag"
)

// A Config specifies the config flags passed to the app
//
// The values set to each flag on init specify the default
// value of each flag
type Config struct {
	Verbose bool // increases logging verbosity

	RequestTimeout time.Duration // the duration to wait before a http transaction completes

	// a file path to save the mirrorlist into
	//
	// Note: "" specifies stdout
	Save    string
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
	//	- [mirrorlist.PrettyJsonFormat]
	Format mirrorlist.FormatOption

	// the max number of enteries in the [mirrorlist.Mirrorlist]
	//
	// Note: 0 indicates unlimited enteries
	Number int

	// NoBenchmark specifies wether a [mirrorlist.Mirrorlist.Benchmark]
	NoBenchmark bool

	mirrorlist.FilterOptions
}

// A Sentinel error that indicates invalid configuration
var ErrInvalidConfig error = errors.New("invalid configuration")

// ConfigError represents an error
// that was encountered when parsing
// the args given to [Config.Parse]
type ConfigError struct {
	strings.Builder
	cause error
}

// Error Returns the a formated string error
func (ce *ConfigError) Error() string {
	return ce.String()
}

// Unwrap returns the cause of the ConfigError
// resulting in the original error to be preserved
func (ce *ConfigError) Unwrap() error {
	return ce.cause
}

// Parse parses args with the fields of c defining the default values
//
// Note: Parse is atomic, meaning that if a parse fails c will not be mutated
func (c *Config) Parse(args []string) error {
	flagSet := flag.NewFlagSet("reflector", flag.ContinueOnError)

	// we create a temp copy so that
	// if the parsing fails due to validation or some other thing,
	// c will not be mutated
	tempC := *c

	// set the verbosity flag
	flagSet.BoolVar(&tempC.Verbose, "verbose", tempC.Verbose, "Increases the logging verbosity")

	// set the benchmark flag
	flagSet.BoolVar(&tempC.NoBenchmark, "no-benchmark", tempC.NoBenchmark, "Skip benchmarking mirrors download speed")

	// set the timeout flag
	flagSet.DurationVar(&tempC.RequestTimeout, "request-timeout", tempC.RequestTimeout, "The duration to wait before a http transaction completes")

	// set the file path flag
	flagSet.StringVar(&tempC.Save, "save", tempC.Save, "The file path to save the output into")

	// set the format flag
	mirrorlistFormat := string(tempC.Format)
	flagSet.StringVar(&mirrorlistFormat, "format", mirrorlistFormat, "The output format")

	// set the sort flag
	sortCriteria := string(tempC.Sort)
	flagSet.StringVar(&sortCriteria, "sort", sortCriteria, "The cirteria to sort the Mirrolist by")

	// set the threads flag
	flagSet.IntVar(&tempC.Threads, "threads", tempC.Threads, "The Maximum number of concurrent workers used to benchmark mirrors download rates (must be greater than 0)")

	// set the filtering flags

	countryFlag := &multivalueflag.MultiValueFlag[string]{DefaultValues: tempC.Countries, FlagName: "country"}
	flagSet.Var(countryFlag, "country", "Only return mirrors belonging to the specified country")

	protocolFlag := &multivalueflag.MultiValueFlag[mirrorlist.ProtocolOption]{DefaultValues: tempC.Protocols, FlagName: "protocol"}
	flagSet.Var(protocolFlag, "protocol", "Only return mirrors belonging to the specified protocol")

	flagSet.DurationVar(&tempC.Age, "age", tempC.Age, "Only return mirrors that have synced at least in the last specified hours")
	flagSet.IntVar(&tempC.Number, "number", tempC.Number, "Only return at most n mirrors")
	flagSet.BoolVar(&tempC.ActiveOnly, "active-only", tempC.ActiveOnly, "Only return mirrors that are active")
	flagSet.Float64Var(&tempC.MinCompletion, "min-completion", tempC.MinCompletion, "Only return mirrors that have at least the specified completion")

	accErr := &ConfigError{}

	flagSet.SetOutput(accErr)

	err := flagSet.Parse(args)

	if err != nil {
		accErr.cause = err
		return accErr
	}

	if flagSet.NArg() > 0 {
		fmt.Fprintf(accErr, "unexpected positional args: %v\n", flagSet.Args())
		accErr.cause = ErrInvalidConfig
		return accErr
	}

	// update tempC.Sort, tempC.Format, tempC.Protocols, and tempC.Countries
	// [strings.ToLower] is used here because the "enum" constants in [mirrorlist] are all lowercased
	tempC.Sort = mirrorlist.SortCriteria(strings.ToLower(sortCriteria))
	tempC.Format = mirrorlist.FormatOption(strings.ToLower(mirrorlistFormat))
	tempC.Protocols = protocolFlag.GetFinalFlagValues()
	tempC.Countries = countryFlag.GetFinalFlagValues()

	tempC.validate(accErr)

	if accErr.Len() > 0 {
		accErr.cause = ErrInvalidConfig
		return accErr
	}

	// mutate c
	*c = tempC

	return nil
}

// validate validates the [Config]
func (c *Config) validate(configError *ConfigError) {

	if c.Threads <= 0 {
		configError.WriteString("Invalid value for option 'threads': value must be greater than 0\n")
	}

	if c.RequestTimeout <= 0 {
		configError.WriteString("Invalid value for option 'request-timeout': value must be greater than 0\n")
	}

	if c.Age < 0 {
		configError.WriteString("Invalid value for option 'age': value must be greater than or equal to 0\n")
	}

	if c.Number < 0 {
		configError.WriteString("Invalid value for option 'number': value must be greater than to or equal to 0\n")
	}

	if c.NoBenchmark && c.Sort == mirrorlist.DownloadRate {
		configError.WriteString("Invalid combination for 'no-benchmark' and 'sort': cannot sort by download-rate when benchmarking is disabled\n")
	}

	if math.IsNaN(c.MinCompletion) || c.MinCompletion < 0 || c.MinCompletion > 100 {
		configError.WriteString("Invalid value for option 'min-completion': value must be between 0 and 100 (inclusive)\n")
	}

	if c.Format != mirrorlist.PacmanFormat && c.Format != mirrorlist.JsonFormat && c.Format != mirrorlist.PrettyJsonFormat {
		fmt.Fprintf(
			configError,
			"Invalid value for option 'format': value must be one of '%s', '%s', or '%s'\n",
			mirrorlist.PrettyJsonFormat,
			mirrorlist.PacmanFormat,
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

	// validate [Config.Protocols]

	for _, protocol := range c.Protocols {
		if protocol != mirrorlist.HTTPSProtocol && protocol != mirrorlist.HTTPProtocol {
			fmt.Fprintf(
				configError,
				"Invalid value for option 'protocol': value must be '%s' or '%s'\n",
				mirrorlist.HTTPSProtocol,
				mirrorlist.HTTPProtocol,
			)

			return
		}
	}
}
