package mirrorlist

import (
	"errors"
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
	MirrorlistFormat FormatOption = "mirrorlist" 
	JsonFormat FormatOption = "json" 
	TableFormat FormatOption = "table" 
)


// Filter filters out mirrors in [Mirrorlist] that do not meet 
// [FilterOptions]
func (ml Mirrorlist) Filter(threadNum int, filterOptions FilterOptions) Mirrorlist{
	//TODO: finish this function

	return nil
}

// Sort sorts the mirrors in [Mirrorlist] based on [SortCriteria]
func (ml Mirrorlist) Sort(criteria SortCriteria) Mirrorlist{

	sortedMirrorlist := make(Mirrorlist, 0, len(ml))

	//TODO: finish this function

	switch criteria { 
		case Score:

		case DownloadRate:

		default:
	}

	return sortedMirrorlist
}

// Format formats a mirrorlist based on the give [FormatOption]
func (ml Mirrorlist) Format(s FormatOption) string{
	// TODO: finish this function

	return ""
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

