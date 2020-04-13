package checker

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type proxyChecker struct {
	locker               sync.Mutex
	busy                 bool
	numWorkers           int
	jobs                 chan proxyCheckRequest
	results              chan proxyCheckResult
	done                 chan bool
	wg                   sync.WaitGroup
	proxyList            []string
	checkURL             string
	tgToken              string
	tgProxy              string
	maxProxiesToCheck    int
	maxAddressesToOutput int
	stats                proxyCheckerStats
	allowedUsers         []string
}

type proxyCheckRequest struct {
	address string
}

type proxyCheckResult struct {
	workerId    int
	address     string
	success     bool
	duration    time.Duration
	errorReason string
}

type proxyCheckerStats struct {
	valid           bool
	startTime       time.Time
	endTime         time.Time
	counterTotal    int
	counterOk       int
	counterFail     int
	checkProgress   float64
	failedAddresses []string
}

func NewProxyChecker(c *ConfigStruct) *proxyChecker {
	pc := proxyChecker{
		numWorkers:           c.NumWorkers,
		jobs:                 nil,
		results:              nil,
		done:                 nil,
		wg:                   sync.WaitGroup{},
		proxyList:            nil,
		checkURL:             c.CheckURL,
		tgToken:              c.TgAPIToken,
		tgProxy:              c.TgProxy,
		maxProxiesToCheck:    c.MaxProxiesToCheck,
		maxAddressesToOutput: c.MaxAddressesToOutput,
		stats:                proxyCheckerStats{valid: false},
	}

	pc.allowedUsers = append(pc.allowedUsers, "plyul")
	return &pc
}

func (pc *proxyChecker) AddProxiesFromFile(filename string) {
	fmt.Printf("Reading proxies from '%s'... ", filename)
	proxies, err := ioutil.ReadFile(filename)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "can't read file with proxies:", err)
		return
	}
	fmt.Println("done")
	pc.proxyList = strings.Split(string(proxies), "\n")
	if len(pc.proxyList) > 0 && pc.proxyList[len(pc.proxyList)-1] == "" {
		pc.proxyList = pc.proxyList[:len(pc.proxyList)-1]
		_, _ = fmt.Fprintln(os.Stderr, "Removed last empty line. That's ok")
	}
	_, _ = fmt.Fprintf(os.Stderr, "Got %d proxies from file (capped on %d)\n", len(pc.proxyList), pc.maxProxiesToCheck)
}

func (pc *proxyChecker) Run() {
	tgAPI := InitTelegramAPI(pc.tgToken, pc.tgProxy)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := tgAPI.GetUpdatesChan(u)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Got error from Telegram bot: %s)\n", err.Error())
		return
	}
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		if !pc.userAllowed(update.Message.From.UserName) {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Отец не разрешает мне разговаривать с незнакомыми.")
			_, _ = tgAPI.Send(msg)
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ParseMode = "MarkdownV2"
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start", "help":
				msg.Text = handleHelp()
			case "getnum":
				msg.Text = pc.handleGetNum()
			case "listall":
				msg.Text = pc.handleListAll()
			case "listfailed":
				msg.Text = pc.handleListFailed()
			case "checkall":
				msg.Text = pc.handleCheck(pc.proxyList)
			case "checkfailed":
				msg.Text = pc.handleCheck(pc.stats.failedAddresses)
			case "stats":
				msg.Text = pc.handleStats()
			default:
				msg.Text = "Не знаю такой команды"
			}
			if msg.Text != "" {
				_, _ = tgAPI.Send(msg)
			}
		} else {
			msg.Text = "Не болтай, работай"
			_, _ = tgAPI.Send(msg)
		}
	}

}

func (pc *proxyChecker) getResults() {
	for result := range pc.results {
		pc.stats.counterTotal++
		if (pc.stats.counterTotal % 10) == 0 {
			b := float64(len(pc.proxyList))
			if pc.maxProxiesToCheck < len(pc.proxyList) {
				b = float64(pc.maxProxiesToCheck)
			}
			pc.stats.checkProgress = float64(pc.stats.counterTotal) / b
		}
		if result.success {
			//fmt.Printf("Proxy %s check OK (checked in %s by worker %d)\n", result.address, result.duration, result.workerId)
			pc.stats.counterOk++
		} else {
			//fmt.Printf("\rProxy %s check FAIL, reason: %s (checked in %s  by worker %d)\n", result.address, result.errorReason, result.duration, result.workerId)
			pc.stats.counterFail++
			pc.stats.failedAddresses = append(pc.stats.failedAddresses, result.address)
		}
	}
	pc.done <- true
}

func (pc *proxyChecker) checkProxyWorker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range pc.jobs {
		result := proxyCheckResult{
			workerId:    id,
			address:     job.address,
			success:     false,
			duration:    0,
			errorReason: "",
		}
		dialer, err := proxy.SOCKS5("tcp", job.address, nil, proxy.Direct)
		if err != nil {
			result.errorReason = fmt.Sprintf("can't connect to the proxy: %s", err)
			pc.results <- result
			continue
		}
		phd := proxy.NewPerHost(dialer, nil)
		httpTransport := &http.Transport{}
		httpClient := &http.Client{Transport: httpTransport}
		httpTransport.DialContext = phd.DialContext
		req, err := http.NewRequest("GET", pc.checkURL, nil)
		if err != nil {
			result.errorReason = fmt.Sprintf("can't create request: %s", err)
			pc.results <- result
			continue
		}
		reqBegin := time.Now()
		resp, err := httpClient.Do(req)
		if err != nil {
			result.errorReason = fmt.Sprintf("can't GET page: %s", err)
			pc.results <- result
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			result.errorReason = fmt.Sprintf("can't read body: %s", err)
			pc.results <- result
			continue
		}
		result.duration = time.Now().Sub(reqBegin)
		bodyString := strings.TrimSpace(string(body))
		resp.Body.Close()
		expected := strings.Split(job.address, ":")[0]
		if bodyString != expected {
			result.errorReason = fmt.Sprintf("expect body '%s', got '%s'", expected, bodyString)
			pc.results <- result
			continue
		}
		result.success = true
		pc.results <- result
	}
}

func (pc *proxyChecker) userAllowed(user string) bool {
	for _, u := range pc.allowedUsers {
		if user == u {
			return true
		}
	}
	return false
}

func (pc *proxyChecker) checkProxies(list []string) {
	pc.locker.Lock()
	if pc.busy == true {
		return
	} else {
		pc.busy = true
	}
	pc.locker.Unlock()

	pc.stats.valid = true
	pc.stats.startTime = time.Now()
	pc.stats.endTime = time.Now()
	pc.stats.failedAddresses = nil
	pc.stats.checkProgress = 0.0
	pc.stats.counterTotal = 0
	pc.stats.counterOk = 0
	pc.stats.counterFail = 0

	// Open channels
	pc.jobs = make(chan proxyCheckRequest)
	pc.results = make(chan proxyCheckResult)
	pc.done = make(chan bool)

	// Start worker goroutines
	fmt.Printf("Creating %d workers... ", pc.numWorkers)
	for id := 1; id <= pc.numWorkers; id++ {
		pc.wg.Add(1)
		go pc.checkProxyWorker(id, &pc.wg)
	}
	fmt.Println("done")

	// Start result receiver goroutine
	go pc.getResults()

	// push jobs to workers
	for n, prx := range list {
		if n > pc.maxProxiesToCheck-1 {
			break
		}
		job := proxyCheckRequest{
			address: prx,
		}
		pc.jobs <- job
	}
	close(pc.jobs)

	// Wait for completion
	pc.wg.Wait()
	close(pc.results)
	pc.stats.checkProgress = 1.0
	<-pc.done
	close(pc.done)

	pc.locker.Lock()
	pc.busy = false
	pc.locker.Unlock()
	pc.stats.endTime = time.Now()
}
