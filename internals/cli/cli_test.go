package cli

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mahdi-lnx/reflector-go/internals/mirrorlist"
)

func TestConfigValidate(t *testing.T){

	newValidConfig := func() Config{
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

	t.Run("reject invalid config", func(t *testing.T) {
		t.Parallel()

		tests := []struct{
			name string
			config Config
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
				wantErrMessageToContain: []string{ "request-timeout", "download-rate", "no-benchmark", "sort", "format", "number" },
			},

		}

		for _, test := range tests{
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				configErr := &ConfigError{}
				config := test.newInvalidConfig()

				config.validate(configErr)

				errMessage := configErr.Error()

				containsFlags := func(errMessage string, flags []string) (ok bool, notContainedFlags string){

					for _, flag := range flags{
						if !strings.Contains(errMessage, flag){
							if notContainedFlags == ""{
								notContainedFlags = fmt.Sprintf("'%s'", flag)
								continue
							}
							notContainedFlags = fmt.Sprintf("%s, '%s'", notContainedFlags, flag)
						}
					}

					if notContainedFlags == ""{
						return true, notContainedFlags
					}

					return false, fmt.Sprintf("[ %s ]", notContainedFlags)
				}
				
				if ok, notContainedFlags := containsFlags(errMessage, test.wantErrMessageToContain); !ok{
					t.Errorf("expected errMessage to contain flags %s, but got `%s`", notContainedFlags, errMessage)
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
			t.Errorf("expected errMessage to be empty but got %s", errMessage)
		}
	})
}

func TestConfigParse(t *testing.T){
	//TODO
}

