package checker

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/proxy"
	"log"
	"net/http"
)

func InitTelegramAPI(apiToken, proxyAddress string) *tgbotapi.BotAPI {
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	var phd *proxy.PerHost
	if proxyAddress != "" {
		dialer, err := proxy.SOCKS5("tcp", proxyAddress, nil, proxy.Direct)
		if err != nil {
			log.Fatal(err)
		}
		phd = proxy.NewPerHost(dialer, nil)
		httpTransport.DialContext = phd.DialContext
	}
	botAPI, err := tgbotapi.NewBotAPIWithClient(apiToken, httpClient)
	if err != nil {
		log.Fatal(err)
	}

	botAPI.Debug = true

	log.Printf("Authorized on account %s", botAPI.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return botAPI
}
