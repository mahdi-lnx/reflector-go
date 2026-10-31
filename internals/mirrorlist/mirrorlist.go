package mirrorlist

import (
	"cmp"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"
)

// Mirror defines a single arch mirror
type Mirror struct{

	// Url defines the url of the Mirror
	Url string `json:"url"`

	// Protocol defines the protocol that the Mirror understands
	Protocol string `json:"protocol"`

	// LastSync defines the datetime that the mirror was last synced in zulu time
	LastSync time.Time `json:"last_sync"`

	Completion float32 `json:"completion_pct"`
	Delay int `json:"delay"`
	DurationAvg float64 `json:"duration_avg"`
	DurationStddev float64 `json:"duration_stddev"`

	// Score defines the Score of the Mirror
	Score int `json:"score"`

	// Active defines wether the mirror is active
	Active bool `json:"active"`

	// Country defines the country name that the mirror is located in
	Country string `json:"country"`

	// CountryCode defines the country code of the country that the mirror is located in
	CountryCode string `json:"country_code"`

	ISOS bool `json:"isos"`

	// IPV4 defines wether the Mirror has a ipv4
	IPV4 bool `json:"ipv4"`


	// IPV6 defines wether the Mirror has a ipv4
	IPV6 bool `json:"ipv6"`

	// DownloadRate defines the downloadrate of the mirror which is calculated dynamically
	// in real-time using [Mirror.Benchmark]
	//
	// Note: arch doesn't return have a "download_rate" field, so this field doesnt have to be a pointer
	DownloadRate float64 `json:"download_rate"` 
}

// Benchmark benchmarks the DownloadRate of a [Mirror] in real-time
func (m Mirror) Benchmark(){
	// TODO: finish this function
}


// Mirrorlist defines a slice of [Mirror]
type Mirrorlist []Mirror


// FilterOption specifies the filters that can be applied to a [Mirrorlist]
type FilterOptions struct{

	// a duration that a [Mirror]'s last sync time must be less than or equal to
	//
	// Note: 0 indicates any sync time is fine
	Age time.Duration 

	// a country that a [Mirror] must belong to
	//
	// Note: an empty string means all countries
	Country string 

	// the protocol that the mirror must support
	//
	// Note: an empty string means any protocol
	Protocol string
}


// SortCriteria defines the sorting criteria of a mirrorlist
type SortCriteria string

const (
	Age SortCriteria = "age" 
	DownloadRate SortCriteria = "download-rate" 
	Score SortCriteria = "score" 
)


// FormatOption defines the formatting of a mirrorlist
type FormatOption string

const (
	PacmanFormat FormatOption = "pacman" 
	JsonFormat FormatOption = "json" 
	TableFormat FormatOption = "table" 
)


// Filter filters out mirrors in [Mirrorlist] that do not meet 
// [FilterOptions]
func (ml Mirrorlist) Filter(filterOptions FilterOptions) Mirrorlist{

	// countryMatches is a helper for filtering mirrors based on their countries
	countryMatches := func(mirror Mirror) bool{
		countryMatches := filterOptions.Country == "" // any country is allowed
		countryMatches = countryMatches || strings.EqualFold(filterOptions.Country, mirror.Country) 
		countryMatches = countryMatches || strings.EqualFold(filterOptions.Country, mirror.CountryCode)

		return countryMatches
	}

	var filteredMirrorlist Mirrorlist
	
	currentTime := time.Now() // calculate the current time once before the filtering begins

	for _, mirror := range ml{

		if !countryMatches(mirror){
			continue
		}

		if filterOptions.Protocol != "" && !strings.EqualFold(mirror.Protocol, filterOptions.Protocol){
			continue
		}

		if filterOptions.Age != 0 && currentTime.Sub(mirror.LastSync) > filterOptions.Age{
			continue
		}

		filteredMirrorlist = append(filteredMirrorlist, mirror)
	}

	return filteredMirrorlist
}

// Sort sorts the mirrors in [Mirrorlist] based on [SortCriteria]
func (ml Mirrorlist) Sort(criteria SortCriteria) (Mirrorlist, error){

	sortedMirrorlist := slices.Clone(ml)

	switch criteria { 
		case Score:
			slices.SortFunc(sortedMirrorlist, func(m1, m2 Mirror) int{
			// if m1 has a lower score, then m1 should be before m2, and vice versa

				return cmp.Compare(m1.Score, m2.Score)
			})
		case DownloadRate:
			slices.SortFunc(sortedMirrorlist, func(m1, m2 Mirror) int{
				// If m1 has a higher donwload rate then, it should be before m2, and vice versa

				// if m2 is bigger than m1, then Compare will return a positive
				// number, meaning that SortFunc would place m1 after m2
				// if m1 is bigger than m2, then Compare will return a negative
				// number, meaning that SortFunc would place m2 after m1

				return cmp.Compare(m2.DownloadRate, m1.DownloadRate)
			})
		case Age: 
			slices.SortFunc(sortedMirrorlist, func(m1, m2 Mirror) int{
				// if m1 has a time that happend after m2, then m1 should be
				// before m2, and vice versa

				// if m2 is before m1, then Compare will return a negative
				// number, meaning that SortFunc would place m2 after m1
				// if m2 is after than m1, then Compare will return a positive
				// number, meaning that SortFunc would place m1 after m2

				return m2.LastSync.Compare(m1.LastSync)
			})
		default:
			return nil, fmt.Errorf("unsupported sort criteria '%s'", criteria)
	}

	return sortedMirrorlist, nil
}

// Format formats a mirrorlist based on the given [FormatOption]
// and writes the output to out
func (ml Mirrorlist) Write(out io.Writer, formatOption FormatOption) error{
	// TODO: finish this function

	switch formatOption{
		case PacmanFormat:
		case TableFormat:
		case JsonFormat: // more explicit than default
			return json.MarshalWrite(out, ml)
	}

	return nil
}

// Limit limits at most n enteries from the start of ml
//
// if n == 0, then a shallow copy of ml is returned
// if n < 0, Limit returns an error
func (ml Mirrorlist) Limit(n int) (Mirrorlist, error){
	if n < 0 {
		return nil, errors.New("n cant not be less than zero")
	}

	// return a shallow copy for all n

	if n == 0{
		return slices.Clone(ml), nil
	}

	return slices.Clone(ml[:min(len(ml), n)]), nil
}

