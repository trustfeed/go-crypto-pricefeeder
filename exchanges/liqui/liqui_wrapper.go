package liqui

import (
	"errors"
	"log"
	"sync"

	"github.com/trustfeed/go-crypto-pricefeeder/common"
	"github.com/trustfeed/go-crypto-pricefeeder/currency/pair"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/orderbook"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/ticker"
)

// Start starts the Liqui go routine
func (l *Liqui) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		l.Run()
		wg.Done()
	}()
}

// Run implements the Liqui wrapper
func (l *Liqui) Run() {
	if l.Verbose {
		log.Printf("%s polling delay: %ds.\n", l.GetName(), l.RESTPollingDelay)
		log.Printf("%s %d currencies enabled: %s.\n", l.GetName(), len(l.EnabledPairs), l.EnabledPairs)
	}

	var err error
	l.Info, err = l.GetInfo()
	if err != nil {
		log.Printf("%s Unable to fetch info.\n", l.GetName())
	} else {
		exchangeProducts := l.GetAvailablePairs(true)
		err = l.UpdateCurrencies(exchangeProducts, false, false)
		if err != nil {
			log.Printf("%s Failed to get config.\n", l.GetName())
		}
	}
}

// UpdateTicker updates and returns the ticker for a currency pair
func (l *Liqui) UpdateTicker(p pair.CurrencyPair, assetType string) (ticker.Price, error) {
	var tickerPrice ticker.Price
	pairsString, err := exchange.GetAndFormatExchangeCurrencies(l.Name,
		l.GetEnabledCurrencies())
	if err != nil {
		return tickerPrice, err
	}

	result, err := l.GetTicker(pairsString.String())
	if err != nil {
		return tickerPrice, err
	}

	for _, x := range l.GetEnabledCurrencies() {
		currency := exchange.FormatExchangeCurrency(l.Name, x).String()
		var tp ticker.Price
		tp.Pair = x
		tp.High = result[currency].High
		tp.Last = result[currency].Last
		tp.Ask = result[currency].Sell
		tp.Bid = result[currency].Buy
		tp.Last = result[currency].Last
		tp.Low = result[currency].Low
		tp.Volume = result[currency].Vol
		ticker.ProcessTicker(l.Name, x, tp, assetType)
	}

	return ticker.GetTicker(l.Name, p, assetType)
}

// GetTickerPrice returns the ticker for a currency pair
func (l *Liqui) GetTickerPrice(p pair.CurrencyPair, assetType string) (ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(l.Name, p, assetType)
	if err != nil {
		return l.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// GetOrderbookEx returns orderbook base on the currency pair
func (l *Liqui) GetOrderbookEx(p pair.CurrencyPair, assetType string) (orderbook.Base, error) {
	ob, err := orderbook.GetOrderbook(l.Name, p, assetType)
	if err != nil {
		return l.UpdateOrderbook(p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (l *Liqui) UpdateOrderbook(p pair.CurrencyPair, assetType string) (orderbook.Base, error) {
	var orderBook orderbook.Base
	orderbookNew, err := l.GetDepth(exchange.FormatExchangeCurrency(l.Name, p).String())
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Bids {
		data := orderbookNew.Bids[x]
		orderBook.Bids = append(orderBook.Bids, orderbook.Item{Amount: data[1], Price: data[0]})
	}

	for x := range orderbookNew.Asks {
		data := orderbookNew.Asks[x]
		orderBook.Asks = append(orderBook.Asks, orderbook.Item{Amount: data[1], Price: data[0]})
	}

	orderbook.ProcessOrderbook(l.Name, p, orderBook, assetType)
	return orderbook.GetOrderbook(l.Name, p, assetType)
}

// GetExchangeAccountInfo retrieves balances for all enabled currencies for the
// Liqui exchange
func (l *Liqui) GetExchangeAccountInfo() (exchange.AccountInfo, error) {
	var response exchange.AccountInfo
	response.ExchangeName = l.GetName()
	accountBalance, err := l.GetAccountInfo()
	if err != nil {
		return response, err
	}

	for x, y := range accountBalance.Funds {
		var exchangeCurrency exchange.AccountCurrencyInfo
		exchangeCurrency.CurrencyName = common.StringToUpper(x)
		exchangeCurrency.TotalValue = y
		exchangeCurrency.Hold = 0
		response.Currencies = append(response.Currencies, exchangeCurrency)
	}

	return response, nil
}

// GetExchangeHistory returns historic trade data since exchange opening.
func (l *Liqui) GetExchangeHistory(p pair.CurrencyPair, assetType string) ([]exchange.TradeHistory, error) {
	var resp []exchange.TradeHistory

	return resp, errors.New("trade history not yet implemented")
}

// SubmitExchangeOrder submits a new order
func (l *Liqui) SubmitExchangeOrder(p pair.CurrencyPair, side string, orderType int, amount, price float64) (int64, error) {
	return 0, errors.New("not yet implemented")
}

// ModifyExchangeOrder will allow of changing orderbook placement and limit to
// market conversion
func (l *Liqui) ModifyExchangeOrder(p pair.CurrencyPair, orderID, action int64) (int64, error) {
	return 0, errors.New("not yet implemented")
}

// CancelExchangeOrder cancels an order by its corresponding ID number
func (l *Liqui) CancelExchangeOrder(p pair.CurrencyPair, orderID int64) (int64, error) {
	return 0, errors.New("not yet implemented")
}

// CancelAllExchangeOrders cancels all orders associated with a currency pair
func (l *Liqui) CancelAllExchangeOrders(p pair.CurrencyPair) error {
	return errors.New("not yet implemented")
}

// GetExchangeOrderInfo returns information on a current open order
func (l *Liqui) GetExchangeOrderInfo(orderID int64) (float64, error) {
	return 0, errors.New("not yet implemented")
}

// GetExchangeDepositAddress returns a deposit address for a specified currency
func (l *Liqui) GetExchangeDepositAddress(p pair.CurrencyPair) (string, error) {
	return "", errors.New("not yet implemented")
}

// WithdrawExchangeFunds returns a withdrawal ID when a withdrawal is submitted
func (l *Liqui) WithdrawExchangeFunds(address string, p pair.CurrencyPair, amount float64) (string, error) {
	return "", errors.New("not yet implemented")
}
