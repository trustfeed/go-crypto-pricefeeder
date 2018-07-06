package main

import (
	"github.com/trustfeed/go-crypto-pricefeeder/communications"
	"github.com/trustfeed/go-crypto-pricefeeder/config"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges"
	"github.com/trustfeed/go-crypto-pricefeeder/portfolio"
)

// Bot contains configuration, portfolio, exchange & ticker data and is the
// overarching type across this code base.
type Bot struct {
	config     *config.Config
	portfolio  *portfolio.Base
	exchanges  []exchange.IBotExchange
	comms      *communications.Communications
	shutdown   chan bool
	dryRun     bool
	configFile string
}