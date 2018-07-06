package bot

import (
	"github.com/trustfeed/go-crypto-pricefeeder/communications"
	"github.com/trustfeed/go-crypto-pricefeeder/config"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges"
	"github.com/trustfeed/go-crypto-pricefeeder/portfolio"
)

// Bot contains configuration, portfolio, exchange & ticker data and is the
// overarching type across this code base.
type Bot struct {
	Config     *config.Config
	Portfolio  *portfolio.Base
	Exchanges  []exchange.IBotExchange
	Comms      *communications.Communications
	Shutdown   chan bool
	DryRun     bool
	ConfigFile string
}