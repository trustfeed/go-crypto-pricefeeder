package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	Bot "github.com/trustfeed/go-crypto-pricefeeder/bot"
	"github.com/trustfeed/go-crypto-pricefeeder/currency"
	"github.com/trustfeed/go-crypto-pricefeeder/currency/forexprovider"

	"github.com/trustfeed/go-crypto-pricefeeder/common"
	"github.com/trustfeed/go-crypto-pricefeeder/communications"
	"github.com/trustfeed/go-crypto-pricefeeder/config"
	"github.com/trustfeed/go-crypto-pricefeeder/portfolio"
)

const banner = `
   ____   ____             ___________ ___.__._______/  |_  ____           
  / ___\ /  _ \   ______ _/ ___\_  __ <   |  |\____ \   __\/  _ \   ______ 
 / /_/  >  <_> ) /_____/ \  \___|  | \/\___  ||  |_> >  | (  <_> ) /_____/ 
 \___  / \____/           \___  >__|   / ____||   __/|__|  \____/          
/_____/                       \/       \/     |__|                         
             .__              _____                .___                    
_____________|__| ____  _____/ ____\____  ____   __| _/___________         
\____ \_  __ \  |/ ___\/ __ \   __\/ __ \/ __ \ / __ |/ __ \_  __ \        
|  |_> >  | \/  \  \__\  ___/|  | \  ___|  ___// /_/ \  ___/|  | \/        
|   __/|__|  |__|\___  >___  >__|  \___  >___  >____ |\___  >__|           
|__|                 \/    \/          \/    \/     \/    \/               
`

var bot Bot.Bot

func main() {
	bot.Shutdown = make(chan bool)
	HandleInterrupt()

	defaultPath, err := config.GetFilePath("")
	if err != nil {
		log.Fatal(err)
	}

	//Handle flags
	flag.StringVar(&bot.ConfigFile, "config", defaultPath, "config file to load")
	dryrun := flag.Bool("dryrun", false, "dry runs bot, doesn't save config file")
	version := flag.Bool("version", false, "retrieves current GoCryptoTrader version")
	flag.Parse()

	if *version {
		fmt.Printf(BuildVersion(true))
		os.Exit(0)
	}

	if *dryrun {
		bot.DryRun = true
	}

	bot.Config = &config.Cfg
	fmt.Println(banner)
	fmt.Println(BuildVersion(false))
	log.Printf("Loading config file %s..\n", bot.ConfigFile)

	err = bot.Config.LoadConfig(bot.ConfigFile)
	if err != nil {
		log.Fatalf("Failed to load config. Err: %s", err)
	}

	AdjustGoMaxProcs()
	log.Printf("Bot '%s' started.\n", bot.Config.Name)
	log.Printf("Bot dry run mode: %v.\n", common.IsEnabled(bot.DryRun))

	log.Printf("Available Exchanges: %d. Enabled Exchanges: %d.\n",
		len(bot.Config.Exchanges),
		bot.Config.CountEnabledExchanges())

	common.HTTPClient = common.NewHTTPClientWithTimeout(bot.Config.GlobalHTTPTimeout)
	log.Printf("Global HTTP request timeout: %v.\n", common.HTTPClient.Timeout)

	SetupExchanges(bot)
	if len(bot.Exchanges) == 0 {
		log.Fatalf("No exchanges were able to be loaded. Exiting")
	}

	log.Println("Starting communication mediums..")
	bot.Comms = communications.NewComm(bot.Config.GetCommunicationsConfig())
	bot.Comms.GetEnabledCommunicationMediums()

	log.Printf("Fiat display currency: %s.", bot.Config.Currency.FiatDisplayCurrency)
	currency.BaseCurrency = bot.Config.Currency.FiatDisplayCurrency
	currency.FXProviders = forexprovider.StartFXService(bot.Config.GetCurrencyConfig().ForexProviders)
	log.Printf("Primary forex conversion provider: %s.\n", bot.Config.GetPrimaryForexProvider())
	err = bot.Config.RetrieveConfigCurrencyPairs(true)
	if err != nil {
		log.Fatalf("Failed to retrieve config currency pairs. Error: %s", err)
	}
	log.Println("Successfully retrieved config currencies.")
	log.Println("Fetching currency data from forex provider..")
	err = currency.SeedCurrencyData(common.JoinStrings(currency.FiatCurrencies, ","))
	if err != nil {
		log.Fatalf("Unable to fetch forex data. Error: %s", err)
	}

	bot.Portfolio = &portfolio.Portfolio
	bot.Portfolio.SeedPortfolio(bot.Config.Portfolio)
	SeedExchangeAccountInfo(GetAllEnabledExchangeAccountInfo().Data)

	go portfolio.StartPortfolioWatcher()
	go TickerUpdaterRoutine()
	go OrderbookUpdaterRoutine()

	if bot.Config.Webserver.Enabled {
		listenAddr := bot.Config.Webserver.ListenAddress
		log.Printf(
			"HTTP Webserver support enabled. Listen URL: http://%s:%d/\n",
			common.ExtractHost(listenAddr), common.ExtractPort(listenAddr),
		)

		router := NewRouter(bot.Exchanges)
		go func() {
			err = http.ListenAndServe(listenAddr, router)
			if err != nil {
				log.Fatal(err)
			}
		}()

		log.Println("HTTP Webserver started successfully.")
		log.Println("Starting websocket handler.")
		StartWebsocketHandler()
	} else {
		log.Println("HTTP RESTful Webserver support disabled.")
	}

	<-bot.Shutdown
	Shutdown()
}

// AdjustGoMaxProcs adjusts the maximum processes that the CPU can handle.
func AdjustGoMaxProcs() {
	log.Println("Adjusting bot runtime performance..")
	maxProcsEnv := os.Getenv("GOMAXPROCS")
	maxProcs := runtime.NumCPU()
	log.Println("Number of CPU's detected:", maxProcs)

	if maxProcsEnv != "" {
		log.Println("GOMAXPROCS env =", maxProcsEnv)
		env, err := strconv.Atoi(maxProcsEnv)
		if err != nil {
			log.Println("Unable to convert GOMAXPROCS to int, using", maxProcs)
		} else {
			maxProcs = env
		}
	}
	if i := runtime.GOMAXPROCS(maxProcs); i != maxProcs {
		log.Fatal("Go Max Procs were not set correctly.")
	}
	log.Println("Set GOMAXPROCS to:", maxProcs)
}

// HandleInterrupt monitors and captures the SIGTERM in a new goroutine then
// shuts down bot
func HandleInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		log.Printf("Captured %v, shutdown requested.", sig)
		bot.Shutdown <- true
	}()
}

// Shutdown correctly shuts down bot saving configuration files
func Shutdown() {
	log.Println("Bot shutting down..")

	if len(portfolio.Portfolio.Addresses) != 0 {
		bot.Config.Portfolio = portfolio.Portfolio
	}

	if !bot.DryRun {
		err := bot.Config.SaveConfig(bot.ConfigFile)

		if err != nil {
			log.Println("Unable to save config.")
		} else {
			log.Println("Config file saved successfully.")
		}
	}

	log.Println("Exiting.")
	os.Exit(0)
}
