package main

import (
	"errors"
	"log"
	"sync"

	Bot "github.com/trustfeed/go-crypto-pricefeeder/bot"
	"github.com/trustfeed/go-crypto-pricefeeder/common"
	exchange "github.com/trustfeed/go-crypto-pricefeeder/exchanges"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/anx"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/binance"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/bitfinex"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/bitflyer"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/bithumb"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/bitstamp"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/bittrex"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/btcc"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/btcmarkets"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/coinut"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/exmo"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/gdax"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/gemini"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/hitbtc"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/huobi"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/itbit"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/kraken"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/lakebtc"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/liqui"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/localbitcoins"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/okcoin"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/okex"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/poloniex"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/wex"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/yobit"
)

// vars related to exchange functions
var (
	ErrNoExchangesLoaded     = errors.New("no exchanges have been loaded")
	ErrExchangeNotFound      = errors.New("exchange not found")
	ErrExchangeAlreadyLoaded = errors.New("exchange already loaded")
	ErrExchangeFailedToLoad  = errors.New("exchange failed to load")
)

// CheckExchangeExists returns true whether or not an exchange has already
// been loaded
func CheckExchangeExists(bot Bot.Bot, exchName string) bool {
	for x := range bot.Exchanges {
		if common.StringToLower(bot.Exchanges[x].GetName()) == common.StringToLower(exchName) {
			return true
		}
	}
	return false
}

// GetExchangeByName returns an exchange given an exchange name
func GetExchangeByName(bot Bot.Bot, exchName string) exchange.IBotExchange {
	for x := range bot.Exchanges {
		if common.StringToLower(bot.Exchanges[x].GetName()) == common.StringToLower(exchName) {
			return bot.Exchanges[x]
		}
	}
	return nil
}

// ReloadExchange loads an exchange config by name
func ReloadExchange(bot Bot.Bot, name string) error {
	nameLower := common.StringToLower(name)

	if len(bot.Exchanges) == 0 {
		return ErrNoExchangesLoaded
	}

	if !CheckExchangeExists(bot, nameLower) {
		return ErrExchangeNotFound
	}

	exchCfg, err := bot.Config.GetExchangeConfig(name)
	if err != nil {
		return err
	}

	e := GetExchangeByName(bot, nameLower)
	e.Setup(exchCfg)
	log.Printf("%s exchange reloaded successfully.\n", name)
	return nil
}

// UnloadExchange unloads an exchange by
func UnloadExchange(bot Bot.Bot, name string) error {
	nameLower := common.StringToLower(name)

	if len(bot.Exchanges) == 0 {
		return ErrNoExchangesLoaded
	}

	if !CheckExchangeExists(bot, nameLower) {
		return ErrExchangeNotFound
	}

	exchCfg, err := bot.Config.GetExchangeConfig(name)
	if err != nil {
		return err
	}

	exchCfg.Enabled = false
	err = bot.Config.UpdateExchangeConfig(exchCfg)
	if err != nil {
		return err
	}

	for x := range bot.Exchanges {
		if bot.Exchanges[x].GetName() == name {
			bot.Exchanges[x].SetEnabled(false)
			bot.Exchanges = append(bot.Exchanges[:x], bot.Exchanges[x+1:]...)
			return nil
		}
	}

	return ErrExchangeNotFound
}

// LoadExchange loads an exchange by name
func LoadExchange(bot Bot.Bot, name string, useWG bool, wg *sync.WaitGroup) error {
	nameLower := common.StringToLower(name)
	var exch exchange.IBotExchange

	if len(bot.Exchanges) > 0 {
		if CheckExchangeExists(bot, nameLower) {
			return ErrExchangeAlreadyLoaded
		}
	}

	switch nameLower {
	case "anx":
		exch = new(anx.ANX)
	case "binance":
		exch = new(binance.Binance)
	case "bitfinex":
		exch = new(bitfinex.Bitfinex)
	case "bitflyer":
		exch = new(bitflyer.Bitflyer)
	case "bithumb":
		exch = new(bithumb.Bithumb)
	case "bitstamp":
		exch = new(bitstamp.Bitstamp)
	case "bittrex":
		exch = new(bittrex.Bittrex)
	case "btcc":
		exch = new(btcc.BTCC)
	case "btc markets":
		exch = new(btcmarkets.BTCMarkets)
	case "coinut":
		exch = new(coinut.COINUT)
	case "exmo":
		exch = new(exmo.EXMO)
	case "gdax":
		exch = new(gdax.GDAX)
	case "gemini":
		exch = new(gemini.Gemini)
	case "hitbtc":
		exch = new(hitbtc.HitBTC)
	case "huobi":
		exch = new(huobi.HUOBI)
	case "itbit":
		exch = new(itbit.ItBit)
	case "kraken":
		exch = new(kraken.Kraken)
	case "lakebtc":
		exch = new(lakebtc.LakeBTC)
	case "liqui":
		exch = new(liqui.Liqui)
	case "localbitcoins":
		exch = new(localbitcoins.LocalBitcoins)
	case "okcoin china":
		exch = new(okcoin.OKCoin)
	case "okcoin international":
		exch = new(okcoin.OKCoin)
	case "okex":
		exch = new(okex.OKEX)
	case "poloniex":
		exch = new(poloniex.Poloniex)
	case "wex":
		exch = new(wex.WEX)
	case "yobit":
		exch = new(yobit.Yobit)
	default:
		return ErrExchangeNotFound
	}

	if exch == nil {
		return ErrExchangeFailedToLoad
	}

	exch.SetDefaults()
	bot.Exchanges = append(bot.Exchanges, exch)
	exchCfg, err := bot.Config.GetExchangeConfig(name)
	if err != nil {
		return err
	}

	exchCfg.Enabled = true
	exch.Setup(exchCfg)

	if useWG {
		exch.Start(wg)
	} else {
		wg := sync.WaitGroup{}
		exch.Start(&wg)
		wg.Wait()
	}
	return nil
}

// SetupExchanges sets up the exchanges used by the bot
func SetupExchanges(bot Bot.Bot) {
	var wg sync.WaitGroup
	for _, exch := range bot.Config.Exchanges {
		if CheckExchangeExists(bot, exch.Name) {
			e := GetExchangeByName(bot, exch.Name)
			if e == nil {
				log.Println(ErrExchangeNotFound)
				continue
			}

			err := ReloadExchange(bot, exch.Name)
			if err != nil {
				log.Printf("ReloadExchange %s failed: %s", exch.Name, err)
				continue
			}

			if !e.IsEnabled() {
				UnloadExchange(bot, exch.Name)
				continue
			}
			return

		}
		if !exch.Enabled {
			log.Printf("%s: Exchange support: Disabled", exch.Name)
			continue
		} else {
			err := LoadExchange(bot, exch.Name, true, &wg)
			if err != nil {
				log.Printf("LoadExchange %s failed: %s", exch.Name, err)
				continue
			}
		}
		log.Printf(
			"%s: Exchange support: Enabled (Authenticated API support: %s - Verbose mode: %s).\n",
			exch.Name,
			common.IsEnabled(exch.AuthenticatedAPISupport),
			common.IsEnabled(exch.Verbose),
		)
	}
	wg.Wait()
}
