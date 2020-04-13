package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"proxychecker/internal/checker"
	"time"
)

func configure() *checker.ConfigStruct {
	const (
		proxyListFileDefault     = "proxies.txt"
		numWorkersDefault        = 100
		maxProxiesToCheckDefault = 15000
		maxAddresssesToOutput    = 50
	)
	var c checker.ConfigStruct
	var needHelp bool
	pflag.BoolVar(&needHelp, "help", false, "Show available configuration options")
	pflag.StringVar(&c.ProxyListFile, "proxy-list", proxyListFileDefault, "File with proxies list")
	pflag.StringVar(&c.CheckURL, "check-url", "", "URL to check through proxies. Must return IP-address of client in response")
	pflag.StringVar(&c.TgAPIToken, "tg-api-token", "", "Telegram Bot API token")
	pflag.StringVar(&c.TgProxy, "tg-proxy-address", "", "Proxy for accessing to Telegram API")
	pflag.IntVar(&c.NumWorkers, "workers", numWorkersDefault, "Number of workers")
	pflag.IntVar(&c.MaxProxiesToCheck, "max-proxies", maxProxiesToCheckDefault, "Maximum number of proxies to check")
	pflag.IntVar(&c.MaxAddressesToOutput, "max-out-addrs", maxAddresssesToOutput, "Maximum number of addresses to output in Telegram bot replies")
	pflag.Parse()
	if needHelp {
		pflag.Usage()
		os.Exit(0)
	}
	if c.CheckURL == "" {
		fmt.Println("Argument --check-url is missing")
		os.Exit(1)
	}
	if c.TgAPIToken == "" {
		fmt.Println("Argument --tg-api-token is missing")
		os.Exit(1)
	}
	return &c
}

func main() {
	c := configure()
	checkBegin := time.Now()

	// do proxy check
	pc := checker.NewProxyChecker(c)
	pc.AddProxiesFromFile(c.ProxyListFile)
	pc.Run()
	fmt.Printf("Check completed in %s\n", time.Now().Sub(checkBegin).String())
}
