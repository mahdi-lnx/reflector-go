package mirrorlist

import (
	"cmp"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
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
	Score float64 `json:"score"`

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

	// DownloadRate defines the download-rate of the mirror which is calculated dynamically
	// in real-time using [Mirror.Benchmark]
	//
	// DownloadRate is calculated as numberOfBytesDownload / elapsedTime (in seconds)
	//
	// Note: arch doesn't return have a "download_rate" field, so this field doesnt have to be a pointer
	DownloadRate float64 `json:"download_rate"` 
}

// Benchmark benchmarks [Mirror.DownloadRate] in real-time
func (m Mirror) Benchmark(ctx context.Context, client *http.Client) (float64, error){

	// set the url to an actual file to download
	url := fmt.Sprintf("%s/core/os/x86_64/core.db", strings.TrimRight(m.Url, "/"),)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil{
		return 0, fmt.Errorf("failed to init a request with context: %w", err)
	}

	startTime := time.Now()

	resp, err := client.Do(req)

	if err != nil{
		return 0, fmt.Errorf("failed to make a get request to '%s': %w", m.Url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		return 0, fmt.Errorf("recieved status code '%s'", resp.Status)
	}

	n, err := io.Copy(io.Discard, resp.Body)

	if err != nil {
		return 0, fmt.Errorf("failed to read the response body: %w", err)
	}

	elapsedTime := time.Since(startTime).Seconds()

	if elapsedTime == 0{ // this is very unlikely to happen
		return 0, fmt.Errorf("invalid calculated elapsed time '%f'", elapsedTime)
	}

	downloadRate := float64(n) / elapsedTime

	return downloadRate, nil
}


// Mirrorlist defines a slice of [Mirror]
type Mirrorlist []Mirror


// FilterOption specifies the filters that can be applied to a [Mirrorlist]
type FilterOptions struct{

	// Age is a duration that a [Mirror]'s last sync time must be less than or equal to
	//
	// Note: 0 indicates any sync time is fine
	Age time.Duration 

	// Country that a [Mirror] must belong to
	//
	// Note: an empty string means all countries
	Country string 

	// Protocol that the mirror must support
	//
	// Note: an empty string means any protocol
	Protocol string

	// MinCompletion specfies the min number that [Mirror.Completion] must be 
	//
	// The value is in precentage, e.g., 99.5 specifies 99.5%
	MinCompletion float32

	// ActiveOnly specifies mirror's that are only active 
	ActiveOnly bool
}


// SortCriteria defines the sorting criteria of a mirrorlist
type SortCriteria string

const (
	Age SortCriteria = "age" // sort by sync time
	DownloadRate SortCriteria = "download-rate"  // sort by download rate
	Score SortCriteria = "score" // sort by score
)


// FormatOption defines the formatting of a mirrorlist
type FormatOption string

const (
	// PacmanFormat is a [FormatOption] thats undrestandble by pacman
	PacmanFormat FormatOption = "pacman" 

	// JsonFormat is a json [FormatOption] thats meant to be used by scripts
	JsonFormat FormatOption = "json"

	// PrettyJsonFormat is a json [FormatOption] thats meant to be readble by humans
	PrettyJsonFormat FormatOption = "pretty-json"
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

		// this check also checks mirros with unknown sync time, e.g, if the mirror's timestamp was set to null
		if filterOptions.Age != 0 && (mirror.LastSync.IsZero() || currentTime.Sub(mirror.LastSync) > filterOptions.Age){
			continue
		}

		if minCompletion := filterOptions.MinCompletion / 100; minCompletion > mirror.Completion{
			continue
		}

		if filterOptions.ActiveOnly && !mirror.Active{
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

	switch formatOption{
		case PacmanFormat:
			// we don't use goroutines because
			// the order in which the mirrors will be writen
			// is not deterministic; hence, the sorting order may not
			// be preserved

			for _, mirror := range ml{
				url := strings.TrimRight(strings.TrimSpace(mirror.Url), "/") 
				
				_, err := fmt.Fprintf(out, "Server = %s/$repo/os/$arch\n", url)
				
				if err != nil{
					return fmt.Errorf("write mirrorlist as pacman format: %w", err)
				}
			}
		case PrettyJsonFormat:
			err := json.MarshalWrite(out, ml, jsontext.Multiline(true))
			if err != nil{
				return fmt.Errorf("write mirrorlist as pretty-json: %w", err)
			}
		case JsonFormat: 
			err := json.MarshalWrite(out, ml)
			if err != nil{
				return fmt.Errorf("write mirrorlist as json: %w", err)
			}
		default:
			return fmt.Errorf("unsupported format option '%s'", formatOption)
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

// Benchmark benchmarks and sets [Mirror.DownloadRate]
// using [Mirror.Benchmark]
//
// Note: Benchmark always spawns threads + 1 goroutines in total, with the extra
// being the error handling goroutine
//
// Note: if ctx is cancelled or its deadline has reached then the error returned by ctx.Err()
// will be an element of the returned slice
func (ml Mirrorlist) Benchmark(ctx context.Context, client *http.Client, threads int) []error{

	var errs []error 

	if threads <= 0 {
		errs = append(errs, errors.New("The number of specified goroutines must be greater than zero"))

		return errs
	}

	mirrorChan:= make(chan *Mirror, 7)
	errChan := make(chan error, 7)
	doneChan := make(chan any)
	var wg sync.WaitGroup


	for range threads{

		wg.Go(func(){
			for{
				select{
					case mirror, ok := <- mirrorChan:
						if !ok || ctx.Err() != nil{
							// if ctx was cancelled or the channel is closed, then dont make a request 
							return
						}

						downloadRate, err := mirror.Benchmark(ctx, client)

						if err != nil{
							if ctx.Err() != nil{
								// error occured due to ctx being cancelled
								// no need to report the error
								return
							}
							errChan <- fmt.Errorf("mirror '%s' caused an error: %w", mirror.Url, err)
							continue
						}

						mirror.DownloadRate = downloadRate
					case <- ctx.Done():
						return
				}
			}
		})
	}

	// spawn an error handling goroutine
	go func(){
		defer close(doneChan)

		for err := range errChan{
			errs = append(errs, err)
		}
	}()

	producerLoop:
	for i := range ml{
		if ctx.Err() != nil{
			// ctx was cancelled, dont send a mirror to the channel
			break 
		}

		select{
			case mirrorChan <- &(ml[i]):

			case <- ctx.Done():
				break producerLoop
		}
	}

	close(mirrorChan)
	wg.Wait()
	close(errChan)
	<- doneChan

	if err := ctx.Err(); err != nil{
		errs = append(errs, err)
	}

	return errs
}

// Fetch fetches a [Mirrorlist] from Arch
func Fetch(ctx context.Context, client *http.Client) (Mirrorlist, error){

	url := "https://archlinux.org/mirrors/status/json/"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to init a request with context: %w", err)
	}

	resp , err := client.Do(req)

	if err != nil{
		return nil, fmt.Errorf("failed to request mirrors from arch: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		return nil, fmt.Errorf("recieved status code '%s'", resp.Status)
	}

	// body is the struct that represents the
	// response's body
	//
	// Note: the actual body returned by arch has other metadata
	// about arch's own mirror sync process
	var body struct{
		Urls Mirrorlist `json:"urls"`
	}

	err = json.UnmarshalRead(resp.Body, &body)

	if err != nil{
		return nil, fmt.Errorf("failed to parse the request body: %w", err)
	}

	return body.Urls, nil
}

