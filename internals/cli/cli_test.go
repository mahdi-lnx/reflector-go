package cli

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mahdi-lnx/reflector-go/internals/mirrorlist"
)

// newValidConfig creates a new valid [Config]
func newValidConfig() Config{
	return Config{
		Save: "",
		Verbose: false,
		Threads: 7,
		RequestTimeout: time.Second * 7,
		NoBenchmark: false,
		Sort: mirrorlist.DownloadRate,
		Age: time.Hour * 24,
		Countries: []string{ "canada", },
		Format: mirrorlist.PacmanFormat,
		Protocol: "https",
		MinCompletion: 100,
		ActiveOnly: true,
	}
}

func TestConfigValidate(t *testing.T){

	t.Run("reject invalid config", func(t *testing.T) {
		t.Parallel()

		tests := []struct{
			name string
			newInvalidConfig func () Config
			wantErrMessageToContain []string
		}{
			{
				name: "reject negative threads",
				newInvalidConfig: func() Config{
					config := newValidConfig()
					config.Threads = -1
					return config
				},
				wantErrMessageToContain: []string{ "threads", },
			},
			{
				name: "reject negative limiting",
				newInvalidConfig: func() Config{
					config := newValidConfig()
					config.Number = -1
					return config
				},
				wantErrMessageToContain: []string{ "number", },
			},
			{
				name: "reject invalid sort criteria",
				newInvalidConfig: func() Config{
					config := newValidConfig()
					config.Sort = "some invalid sort criteria"
					return config
				},
				wantErrMessageToContain: []string{ "sort", },
			},
			{
				name: "reject invalid output format",
				newInvalidConfig: func() Config{
					config := newValidConfig()
					config.Format = "some invalid output format"
					return config
				},
				wantErrMessageToContain: []string{ "format", },
			},
			{
				name: "reject download sorting when benchmarking is disabled",
				newInvalidConfig: func() Config{
					config := newValidConfig()
					config.Sort = mirrorlist.DownloadRate
					config.NoBenchmark = true
					return config
				},
				wantErrMessageToContain: []string{ "no-benchmark", "sort", },
			},
			{
				name: "reject invalid request-timeout",
				newInvalidConfig: func() Config{
					config := newValidConfig()
					config.RequestTimeout = 0
					return config
				},
				wantErrMessageToContain: []string{ "request-timeout", },
			},
			{
				name: "reject multiple invalid flags",
				newInvalidConfig: func() Config{
					config := newValidConfig()
					config.Sort = mirrorlist.DownloadRate
					config.NoBenchmark = true
					config.Format = "some invalid output format"
					config.Number = -1
					config.RequestTimeout = -1
					config.Threads = 0
					return config
				},
				wantErrMessageToContain: []string{ 
					"threads", "request-timeout", "download-rate", 
					"no-benchmark", "sort", "format", "number",
				},
			},

		}

		for _, test := range tests{
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				configErr := &ConfigError{}
				config := test.newInvalidConfig()

				config.validate(configErr)

				errMessage := configErr.Error()

				var missingTokens []string 

				for _, expectedToken := range test.wantErrMessageToContain{
					if !strings.Contains(errMessage, expectedToken){
						missingTokens = append(missingTokens, expectedToken)
					}
				}

				if len(missingTokens) > 0{
					// the errMessage did not contain an expected token
					t.Errorf(
						"expected errMessage to contain flags [ %s ], but got %q", 
						strings.Join(missingTokens, ", "), 
						errMessage,
					)
				}
			})
		}
	})

	t.Run("accept valid config", func(t *testing.T) {
		t.Parallel()

		config := newValidConfig()
		configErr := &ConfigError{}

		config.validate(configErr)

		errMessage := configErr.Error()

		if errMessage != ""{
			t.Errorf("expected errMessage to be empty, got %q", errMessage)
		}
	})
}

func TestConfigParse(t *testing.T){

	type TestCase struct{
		name string
		flags []string
		check func (flags []string) error
	}

	t.Parallel()

	t.Run("reject invalid flags", func(t *testing.T) {
		t.Parallel()

		tests := []TestCase{
			{
				name: "reject unknown flags",
				flags: []string{ "--some unknown flag", "some unknown value for the unknown flag" },
				check: func(flags []string) error{
					config := newValidConfig()

					err := config.Parse(flags)

					if err == nil{
						return errors.New("expected err to not be nil after parsing unknown flags")
					}

					if errors.Is(err, ErrInvalidConfig){ 
						return fmt.Errorf("expected err's error tree to not have ErrInvalidConfig")
					}

					if defaultConfig := newValidConfig(); !reflect.DeepEqual(config, defaultConfig){
						return fmt.Errorf("expected config to be: %+v\ngot: %+v", defaultConfig, config)
					}

					return nil
				},
			},
			{
				name: "reject invalid age filtering",
				flags: []string{ "--age", "-2h" },
				check: func(flags []string) error{
					config := newValidConfig()

					err := config.Parse(flags)

					if err == nil{
						return errors.New("expected err to not be nil after parsing unknown flags")
					}

					if !errors.Is(err, ErrInvalidConfig){ 
						return fmt.Errorf("expected err's error tree to have ErrInvalidConfig")
					}

					if defaultConfig := newValidConfig(); !reflect.DeepEqual(config, defaultConfig){
						return fmt.Errorf("expected config to be: %+v\ngot: %+v", defaultConfig, config)
					}

					return nil
				},
			},
			{
				name: "reject invalid sort criteria",
				flags: []string{ "--sort", "some invalid sort criteria" },
				check: func(flags []string) error{
					config := newValidConfig()

					err := config.Parse(flags)

					if err == nil{
						return errors.New("expected err to not be nil after parsing unknown flags")
					}

					if !errors.Is(err, ErrInvalidConfig){ 
						return fmt.Errorf("expected err's error tree to have ErrInvalidConfig")
					}

					if defaultConfig := newValidConfig(); !reflect.DeepEqual(config, defaultConfig){
						return fmt.Errorf("expected config to be: %+v\ngot: %+v", defaultConfig, config)
					}

					return nil
				},
			},
			{
				name: "reject invalid limits",
				flags: []string{ "--number", "-1" },
				check: func(flags []string) error{
					config := newValidConfig()

					err := config.Parse(flags)

					if err == nil{
						return errors.New("expected err to not be nil after parsing unknown flags")
					}

					if !errors.Is(err, ErrInvalidConfig){ 
						return fmt.Errorf("expected err's error tree to have ErrInvalidConfig")
					}

					if defaultConfig := newValidConfig(); !reflect.DeepEqual(config, defaultConfig){
						return fmt.Errorf("expected config to be: %+v\ngot: %+v", defaultConfig, config)
					}

					return nil
				},
			},
			{
				name: "reject invalid limits",
				flags: []string{ "--number", "-1" },
				check: func(flags []string) error{
					config := newValidConfig()

					err := config.Parse(flags)

					if err == nil{
						return errors.New("expected err to not be nil after parsing unknown flags")
					}

					if !errors.Is(err, ErrInvalidConfig){ 
						return fmt.Errorf("expected err's error tree to have ErrInvalidConfig")
					}

					if defaultConfig := newValidConfig(); !reflect.DeepEqual(config, defaultConfig){
						return fmt.Errorf("expected config to be: %+v\ngot: %+v", defaultConfig, config)
					}

					return nil
				},
			},
			{
				name: "reject invalid threads",
				flags: []string{ "--threads", "-1" },
				check: func(flags []string) error{
					config := newValidConfig()

					err := config.Parse(flags)

					if err == nil{
						return errors.New("expected err to not be nil after parsing unknown flags")
					}

					if !errors.Is(err, ErrInvalidConfig){ 
						return fmt.Errorf("expected err's error tree to have ErrInvalidConfig")
					}

					if defaultConfig := newValidConfig(); !reflect.DeepEqual(config, defaultConfig){
						return fmt.Errorf("expected config to be: %+v\ngot: %+v", defaultConfig, config)
					}

					return nil
				},
			},
			{
				name: "reject invalid combonations of flags",
				flags: []string{ 
					"--sort", "some invalid sort criteria", 
					"--country", "2",
					"--country", "3",
					"--age", "-2h",
					"--number", "0",
					"--threads", "0",
					"--request-timeout", "0",
				},
				check: func(flags []string) error{
					config := newValidConfig()

					err := config.Parse(flags)

					if err == nil{
						return errors.New("expected err to not be nil after parsing invalid a combo of flags")
					}

					if !errors.Is(err, ErrInvalidConfig){ 
						return fmt.Errorf("expected err's error tree to contain the ErrInvalidConfig sentinel error")
					}

					if defaultConfig:= newValidConfig(); !reflect.DeepEqual(config, defaultConfig){
						return fmt.Errorf("expected config to be: %+v\ngot: %+v", defaultConfig, config)
					}

					return nil
				},
			},
		}

		for _, test := range tests{
			t.Run(test.name, func(t *testing.T){
				if err := test.check(test.flags); err != nil{
					t.Error(err)
				}
			})
		}
	})

	t.Run("parse valid flags", func (t *testing.T){
		t.Parallel()

		tests := []TestCase{
			{
				name: "parse --help",
				flags: []string{ "--help" },
				check: func(flags []string) error{
					config := newValidConfig()

					err := config.Parse(flags)
					
					if err == nil || !errors.Is(err, flag.ErrHelp){
						return errors.New("expected some error in err's tree match flag.ErrHelp")
					}

					if defaultConfig := newValidConfig(); !reflect.DeepEqual(config, defaultConfig){
						return fmt.Errorf(
							"config was modified after parsing --help:\nwant%+v\ngot: %+v", 
							defaultConfig,
							config,
						)
					}

					return nil
				},
			},
			{
				name: "accept default flag values",
				flags: []string{},
				check: func (flags []string) error{
					config := newValidConfig()

					err := config.Parse(flags)

					if err != nil{
						return fmt.Errorf("expected err to be nil, but got: %w", err)
					}

					if defaultConfig := newValidConfig(); !reflect.DeepEqual(config, defaultConfig){
						return fmt.Errorf(
							"config was modified after parsing default values:\nwant%+v\ngot: %+v", 
							defaultConfig,
							config,
						)
					}

					return nil
				},
			},
			{
				name: "accept valid flags",
				flags: []string{ "--country=canada,canada", "--sort", string(mirrorlist.Score), "--format", string(mirrorlist.JsonFormat), },
				check: func(flags []string) error{
					config := newValidConfig()

					err := config.Parse(flags)

					if err != nil{
						return fmt.Errorf("expected err to be nil, but got: %w", err)
					}

					newConfig := newValidConfig()
					newConfig.Countries = []string{ "canada", "canada" }
					newConfig.Sort = mirrorlist.Score
					newConfig.Format = mirrorlist.JsonFormat

					if !reflect.DeepEqual(config, newConfig){
						return fmt.Errorf(
							"parsed config does not match expected config:\nwant:%+v\ngot:%+v", 
							newConfig, 
							config,
						)
					}

					return nil
				},
			},
		}

		for _, test := range tests{
			t.Run(test.name, func(t *testing.T) {
				if err := test.check(test.flags); err != nil{
					t.Error(err)
				}
			})
		}
	})
}

