package main

import (
	"fmt"
	"github.com/rverst/go-miab/miab"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Provides a simple method to update custom address records on your Mail-in-a-Box instance. This tool is designed to
// run e.g. in a docker container and updates the address records for the configured domains at a given interval.
func main() {

	fmt.Println("dnsudpate version: 1.0.0")
	interval, err := strconv.ParseInt(os.Getenv("DNS_INTERVAL"), 10, 32)
	if err != nil {
		fmt.Println("Unable to parse interval (seconds), check environment (DNS_INTERVAL). Has to be a unsigned int.")
		os.Exit(99)
	}

	if interval < 30 {
		interval = 30
	}

	user := os.Getenv("DNS_USER")
	pass := os.Getenv("DNS_PASSWORD")
	endp := os.Getenv("DNS_ENDPOINT")

	c, err := miab.NewConfig(user, pass, endp)
	if err != nil {
		fmt.Println(err)
		os.Exit(91)
	}

	doV4, err := strconv.ParseBool(os.Getenv("DNS_A"))
	if err != nil {
		doV4 = false
	}
	doV6, err := strconv.ParseBool(os.Getenv("DNS_AAAA"))
	if err != nil {
		doV6 = false
	}
	d := os.Getenv("DNS_DOMAINS")
	var domains []string
	if strings.Contains(d, ";") {
		domains = strings.Split(d, ";")
	} else {
		domains = strings.Split(d, ",")
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	defer ticker.Stop()

	go func() {
		<-sigs         // blocked until signal received
		ticker.C = nil // close ticker channel
		done <- true   // write on channel to release
	}()

	go func() {
		doDnsUpdate(c, domains, doV4, doV6)
		for range ticker.C {
			doDnsUpdate(c, domains, doV4, doV6)
		}
	}()

	// blocked until channel is written
	<-done
}

func doDnsUpdate(config *miab.Config, domains []string, doV4 bool, doV6 bool) {

	for _, d := range domains {
		d = strings.Trim(d, " ")
		if d == "" {
			continue
		}

		if doV4 {
			b, err := miab.UpdateDns4(config, d, "")
			if err != nil {
				fmt.Printf("DNS update (A) for '%s' failed with error: %v\n", d, err)
			} else if b {
				fmt.Printf("DNS update (A) for '%s' successful\n", d)
			} else {
				fmt.Printf("DNS update (A) for '%s' failed\n", d)
			}

		}
		if doV6 {
			b, err := miab.UpdateDns6(config, d, "")
			if err != nil {
				fmt.Printf("DNS update (AAAA) for '%s' failed with error: %v\n", d, err)
			} else if b {
				fmt.Printf("DNS update (AAAA) for '%s' successful\n", d)
			} else {
				fmt.Printf("DNS update (AAAA) for '%s' failed\n", d)
			}
		}
	}

}
