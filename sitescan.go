// sitescan is a basic web scraping tool that compares two file trees, and
// prints out the differences.
//
// sitescan can be configured in several different ways. At a minimum, it needs
// to be told the correct URLs to visit. It can also handle basic HTTP authentication
// (username and password). Optionally, you can specify a friendlier name for
// each site, as well. Because it uses Viper for configuration processing, sitescan
// is very flexible in terms of how to configure it. It will accept a YAML based
// command line options, environment variables, and config files - or a combination of
// all three. Precedence is as listed.
//
//
// Command Line Usage:
//
// 	 -c, --config string      path to alternate configuration file
//       --site1 string       Site 1 URL
//       --site1name string   Site 1 Name
//       --site1pass string   Site 1 Password
//       --site1user string   Site 1 User ID
//       --site2 string       Site 2 URL
//       --site2name string   Site 2 Name
//       --site2pass string   Site 2 Password
//       --site2user string   Site 2 User ID
//
//
// Environment Variables
//
// Acceptable environment variables are all capitals, are prefixed with "SITESCAN_",
// and otherwise match the command line switches:
//
//	SITESCAN_SITE1
//	SITESCAN_SITE1NAME
//	SITESCAN_SITE1PASS
//	SITESCAN_SITE1USER
//	SITESCAN_SITE2
//	SITESCAN_SITE2NAME
//	SITESCAN_SITE2PASS
//	SITESCAN_SITE2USER
//
//
// Config File
//
// The default configuration file is named "sitescan_config.yaml" and should reside
// in the directory you're running sitescan from (i.e. the directory that sitescan
// will see as "PWD"). You can specify an alternate config file name/path using the
// -c / --config command line option. And example config file:
// 	# Example sitescan_config.yaml file
// 	site1: http://webserver.myhost.com/path/to/examine
// 	site2: http://www.anotherhost.org:8080/
// 	site1user: someguy
// 	site1pass: spaceballs12345
// 	site1name: MyHost.com site
// 	# site2user:
// 	# site2pass:
// 	site2name: AnotherHost site
//

package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/davexre/syncedData"
	"github.com/gosuri/uilive"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var site1Map = make(map[string]string)
var site2Map = make(map[string]string)

var updateInterval = time.Millisecond * 200

var site1done, site2done, stopupdating chan bool
var site1Counter, site2Counter syncedData.Counter

var lw = uilive.New()

var url1, url2 string
var site1User, site1Pass, site1Name string
var site2User, site2Pass, site2Name string

// these are various anchor texts that are presented by the web browser that
// change sort order, or take us up a directory, etc. We don't want to take
// these into account in our Maps, so we use this list to ignore them when
// we build the maps.
var ignoreThese = map[string]int{
	"Name":             1,
	"Last modified":    2,
	"Size":             3,
	"Description":      4,
	"Parent Directory": 5,
	"Type":             6,
	"..":               7,
	"../":              8,
}

var wg sync.WaitGroup

func config() {

	var clConfigFile, clConfigFileFSName string
	var flagSite1, flagSite1User, flagSite1Pass, flagSite1Name string
	var flagSite2, flagSite2User, flagSite2Pass, flagSite2Name string
	var err error

	v := viper.New()
	flag.StringVarP(&clConfigFile, "config", "c", "", "path to alternate configuration file")
	flag.StringVar(&flagSite1, "site1", "", "Site 1 URL")
	flag.StringVar(&flagSite1User, "site1user", "", "Site 1 User ID")
	flag.StringVar(&flagSite1Pass, "site1pass", "", "Site 1 Password")
	flag.StringVar(&flagSite1Name, "site1name", "", "Site 1 Name")
	flag.StringVar(&flagSite2, "site2", "", "Site 2 URL")
	flag.StringVar(&flagSite2User, "site2user", "", "Site 2 User ID")
	flag.StringVar(&flagSite2Pass, "site2pass", "", "Site 2 Password")
	flag.StringVar(&flagSite2Name, "site2name", "", "Site 2 Name")
	flag.Parse()

	if clConfigFile != "" {
		if strings.HasSuffix(clConfigFile, ".yaml") {
			clConfigFileFSName = clConfigFile
			clConfigFile = strings.TrimSuffix(clConfigFile, ".yaml")
		} else {
			clConfigFileFSName = fmt.Sprintf("%s.yaml", clConfigFile)
		}

		if _, err = os.Stat(clConfigFileFSName); err != nil {
			fmt.Println("config file not found: ", clConfigFileFSName)
			v.SetConfigName("sitescan_config")
		} else {
			v.SetConfigName(clConfigFile)
		}
	} else {
		v.SetConfigName("sitescan_config")
	}

	v.SetDefault("site1", "http://127.0.0.1")
	v.SetDefault("site1user", "")
	v.SetDefault("site1pass", "")
	v.SetDefault("site1name", "Site 1")
	v.SetDefault("site2", "http://127.0.0.1")
	v.SetDefault("site2user", "")
	v.SetDefault("site2pass", "")
	v.SetDefault("site2name", "Site 2")
	v.SetEnvPrefix("SITESCAN")
	v.AutomaticEnv()
	v.BindPFlags(flag.CommandLine)
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("config file not found (viper)\n")
		} else {
			fmt.Printf("%v\n", err)
		}
	}

	url1 = v.GetString("site1")
	url2 = v.GetString("site2")
	site1User = v.GetString("site1user")
	site1Pass = v.GetString("site1pass")
	site1Name = v.GetString("site1name")
	site2User = v.GetString("site2user")
	site2Pass = v.GetString("site2pass")
	site2Name = v.GetString("site2name")

}

// walkLink builds a map of the URLs and plain text names for all the files
// stored at the indicated site. This is intended to be called in a recursive
// fashion between two different goroutines.
//
// So, why use the anchor tag text, and why are we checking the URL in href for
// a trailing slash? Different web servers encode data differently, and present
// the text in the anchor tags differently. For instance, lighthttpd does not
// include the trailing "/" in the anchor tag text, but apache does. lighthttpd
// encodes apostrophes (%27), but apache leaves them as bare apostrophes.
//
// Using anchor tag text skips most of the problems with encoding, but makes it
// hard to recognize directories - and also makes it hard to directly compare
// between different servers. Thus, we check the href URL for a trailing slash,
// and append one to the anchor tag text if it's not already there
func walkLink(urlprefix string, url string, currentName string, siteMap *map[string]string,
	user string, pass string, counter *syncedData.Counter) {

	client := &http.Client{}

	urltoget := fmt.Sprintf("%s%s", urlprefix, url)

	req, _ := http.NewRequest("GET", urltoget, nil)
	if user != "" || pass != "" {
		req.SetBasicAuth(user, pass)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close() // ensure this gets garbage collected at some point

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// we expect to find a file listing at the target URL. We're going to skip over
	// the headers, and process the file and directory links. If it's a directory, we
	// need to dive into it and find anything else
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		_, exists := ignoreThese[s.Text()]
		if !exists {
			href, exists := s.Attr("href")
			if exists {

				// update the progress bar counter
				counter.Incr()

				ourname := fmt.Sprintf("%s%s", currentName, s.Text())
				oururl := fmt.Sprintf("%s%s", url, href)

				// note: maps are not concurrency safe... but we shouldn't be
				// accessing our maps concurrently, so shouldn't be a problem?
				// we can wrap the map in a Mutex, if we need to, but
				// seems like overkill?

				if strings.HasSuffix(href, "/") && !strings.HasSuffix(ourname, "/") {
					ourname = fmt.Sprintf("%s/", ourname)
				}

				(*siteMap)[ourname] = oururl

				if strings.HasSuffix(href, "/") {
					walkLink(urlprefix, oururl, ourname, siteMap, user, pass, counter)
				}

			}

		}

	})

}

func walkLinkWrapper(urlprefix string, currentName string, siteMap *map[string]string,
	user, pass string, done chan bool, counter *syncedData.Counter) {

	walkLink(urlprefix, "", "", siteMap, user, pass, counter)
	done <- true
	wg.Done()

}

func updateProgress() {

	startTime := time.Now()
	var s1Duration, s2Duration time.Duration

	s1done := false
	s2done := false

	for {
		select {
		case <-time.After(updateInterval):
			if !s1done {
				s1Duration = time.Since(startTime)
			}

			fmt.Fprintf(lw, "%-20s %-6s %5v files and directories", site1Name+":",
				s1Duration.Round(time.Second).String(), site1Counter.Read())

			if s1done {
				fmt.Fprintf(lw, " - DONE!\n")
			} else {
				fmt.Fprintf(lw, "\n")
			}

			if !s2done {
				s2Duration = time.Since(startTime)
			}

			fmt.Fprintf(lw.Newline(), "%-20s %-6s %5v files and directories", site2Name+":",
				s2Duration.Round(time.Second).String(), site2Counter.Read())

			if s2done {
				fmt.Fprintf(lw, " - DONE!\n")
			} else {
				fmt.Fprintf(lw, "\n")
			}

		case s1done = <-site1done:
			s1Duration = time.Since(startTime)

		case s2done = <-site2done:
			s2Duration = time.Since(startTime)

		case <-stopupdating:
			fmt.Fprintf(lw, "%-20s %-6s %5v files and directories - DONE!\n", site1Name+":",
				s1Duration.Round(time.Second).String(), site1Counter.Read())
			fmt.Fprintf(lw.Newline(), "%-20s %-6s %5v files and directories - DONE!\n", site2Name+":",
				s2Duration.Round(time.Second).String(), site2Counter.Read())

			lw.Stop()

			return
		}
	}
}

func compareMaps(sm1, sm2 *map[string]string, sitename string) {

	banner := "Files/directories only at "
	fmt.Printf("%s%s:\n", banner, sitename)
	for i := 0; i < len(banner+sitename+":"); i++ {
		fmt.Printf("=")
	}
	fmt.Printf("\n\n")

	// alpha sort the keys

	keys := make([]string, 0, len(*sm1))
	for k := range *sm1 {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		_, exists := (*sm2)[k]
		if !exists {
			fmt.Println(k)
		}
	}

	fmt.Printf("\n\n")

}

func main() {

	config()

	if url1 == url2 {
		fmt.Printf("Both sites are the same:\n")
		fmt.Printf("    Site 1: %s\n", url1)
		fmt.Printf("    Site 2: %s\n\n", url2)
		fmt.Printf("Nothing to compare...")
		os.Exit(1)
	}

	fmt.Println("")
	fmt.Printf("%-20s %s\n", site1Name+":", url1)
	fmt.Printf("%-20s %s\n", site2Name+":", url2)

	fmt.Printf("\nConnecting to servers...\n\n")

	site1done = make(chan bool)
	site2done = make(chan bool)
	stopupdating = make(chan bool)

	lw.Start()

	wg.Add(1)
	go walkLinkWrapper(url1, "", &site1Map, site1User, site1Pass, site1done, &site1Counter)

	wg.Add(1)
	go walkLinkWrapper(url2, "", &site2Map, site2User, site2Pass, site2done, &site2Counter)

	go updateProgress()

	wg.Wait()

	stopupdating <- true

	// this pause helps prevent a case where the updateProgress thread doesn't get
	// finished before we start writing to the screen, so the output looks odd. Rather
	// than add a second waitqueue, this seemed like a more reasonable approach.
	time.Sleep(time.Second)

	fmt.Printf("\n\n")

	compareMaps(&site1Map, &site2Map, site1Name)
	compareMaps(&site2Map, &site1Map, site2Name)

}
