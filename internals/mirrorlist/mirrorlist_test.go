package mirrorlist

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

var client *http.Client
var ctx context.Context
var mr Mirrorlist

func TestMain(m *testing.M){

	client = &http.Client{
		Timeout: time.Second * 7,
	}
	ctx = context.TODO()

	mr2, err := Fetch(ctx, client)

	if err != nil{
		fmt.Println(fmt.Errorf("coult not generate the mirrorlist: %w", err))
		os.Exit(1)
	}

	if len(mr2) == 0{
		fmt.Println(fmt.Errorf("expected len(mirrorlist) > 0"))
		os.Exit(1)
	}

	mr = mr2

	code := m.Run()

	os.Exit(code)
}

func TestMirrorlistFilter(t *testing.T){

	tests := []struct{
		name string
		check func() error
	}{
		{
			name: "filter by country name",
			check: func() error{
				filterOptions := FilterOptions{ Countries: []string{ "CANADA", "England", }, }
				filterdMirrorlist := mr.Filter(filterOptions)

				for _, m := range filterdMirrorlist{
					if country:= strings.ToUpper(m.Country); !slices.Contains(filterOptions.Countries, country){
						return errors.New("expected all mirrors to belong to [ 'CANADA', 'England' ]")
					}
				}

				return nil
			},
		},
		{
			name: "filter by country code",
			check: func() error{
				filterOptions := FilterOptions{ Countries: []string{ "CAD", "UK", }, }
				filterdMirrorlist := mr.Filter(filterOptions)

				for _, m := range filterdMirrorlist{
					if country:= strings.ToUpper(m.CountryCode); !slices.Contains(filterOptions.Countries, country){
						return errors.New("expected all mirrors to belong to [ 'CAD', 'UK' ]")
					}
				}
				return nil
			},
		},
		{
			name: "filter by protocol",
			check: func() error{
				filterOptions := FilterOptions{ Protocol: "https", }
				filterdMirrorlist := mr.Filter(filterOptions)

				for _, m := range filterdMirrorlist{
					if m.Protocol != filterOptions.Protocol{
						return errors.New("expected mirror's protocol to be https")
					}
				}

				return nil
			},
		},
		{
			name: "filter by activity",
			check: func() error{
				filterOptions := FilterOptions{ ActiveOnly: true, }
				filterdMirrorlist := mr.Filter(filterOptions)

				for _, m := range filterdMirrorlist{
					if !m.Active{
						return errors.New("expected all mirrors to be active")
					}
				}

				return nil
			},
		},
		{
			name: "filter by sync time",
			check: func() error{
				filterOptions := FilterOptions{ Age: time.Hour * 24, }
				filterdMirrorlist := mr.Filter(filterOptions)

				for _, m := range filterdMirrorlist{
					if time.Since(m.LastSync) > filterOptions.Age{
						return errors.New("expected all mirrors to have been last sync in the past 24h")
					}
				}

				return nil
			},
		},
		{
			name: "filter by min-completion",
			check: func() error{
				filterOptions := FilterOptions{ MinCompletion: 100}
				filterdMirrorlist := mr.Filter(filterOptions)

				for _, m := range filterdMirrorlist{
					if m.Completion * 100 != filterOptions.MinCompletion{
						return fmt.Errorf("expected all mirrors to have a min-completion of %f%%", filterOptions.MinCompletion)
					}
				}

				return nil
			},
		},
		{
			name: "filter by multiple options",
			check: func() error{
				filterOptions := FilterOptions{ 
					MinCompletion: 100,
					Countries: []string { "canada", },
					Protocol: "https",
					ActiveOnly: true,
				}
				filterdMirrorlist := mr.Filter(filterOptions)

				var (

					errs []error
					minCompletionFlag bool
					protocolFlag bool
					countriesFlag bool
					activeOnlyFlag bool
				)

				for _, m := range filterdMirrorlist{
					if !protocolFlag && m.Protocol != filterOptions.Protocol{
						errs = append(errs, fmt.Errorf("expected all mirrors to be of the %s protocol", filterOptions.Protocol))
						protocolFlag = true
					}


					if !activeOnlyFlag && !m.Active{
						errs = append(errs, errors.New("expected all to be active"))
						activeOnlyFlag = true
					}

					if !minCompletionFlag && m.Completion * 100 != filterOptions.MinCompletion{
						errs = append(errs, fmt.Errorf("expected all mirrors to have a min-completion of %f%%", filterOptions.MinCompletion))
						minCompletionFlag = true
					}

					if !countriesFlag && !slices.Contains(filterOptions.Countries, strings.ToLower(m.Country)){
						errs = append(errs, errors.New("expected all mirrors to belong to [ 'canada' ]"))
						countriesFlag = true
					}

					if minCompletionFlag && protocolFlag && activeOnlyFlag && countriesFlag{
						break
					}
				}

				if errs == nil{
					return nil
				}

				return errors.Join(errs...)
			},
		},
	}
	
	for _, test := range tests{
		t.Run(test.name, func(t *testing.T){
			t.Parallel() // this is safe because Mirrorlist.Filter does not modify Mirrorlist

			if err := test.check(); err != nil{
				t.Error(err)
			}
		})
	}
}

