
package mirrorlist

import (
	"strings"
	"time"
)


// Mirror defines a single arch mirror
type Mirror struct{
	Url string
	Protocol string
	LastSync time.Time
	Completion float32
	Delay int
	DurationAvg float64
	DurationStddev float64
	Score int
	Active bool
	Country string
	CountryCode string
	ISOS bool
	IPV4 bool
	IPV6 bool
	DownloadRate string
}

// SetDownloadRate sets the DownloadRate of a [Mirror]
// in real-time
func (m *Mirror) SetDownloadRate(){
	// TODO: set the downloadRate of m
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

	// the number of enteries in the [Mirrorlist]
	//
	// Note: 0 indicates unlimited enteries
	Number int 
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
func (ml Mirrorlist) Filter(filterOptions FilterOptions) Mirrorlist{

	filterdMirrorlist := make(Mirrorlist, 0, len(ml))

	currentTime := time.Now() // capture the current time before the loop starts

	for _, mirror := range ml{

		if filterOptions.Number != 0 && len(filterdMirrorlist) >= filterOptions.Number{
			// filteredMirrorlist has reached config.Number
			break
		}

		include := true

		// check if config.Country is equal to mirror.Country
		// or mirror.CountryCode using case-insensitivity
		if !strings.EqualFold(mirror.Country, filterOptions.Country){
			include = false
		} else if !strings.EqualFold(mirror.CountryCode, filterOptions.Country){
			include = false
		}

		// check if the mirror was last sync before than
		// the specifed config.Age
		if filterOptions.Age != 0 && currentTime.Sub(mirror.LastSync) > filterOptions.Age{
			include = false
		}

		if !include {
			// the mirror is filtered out
			continue
		}

		// mirror passed the filter
		filterdMirrorlist = append(filterdMirrorlist, mirror)
	}

	return filterdMirrorlist
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

