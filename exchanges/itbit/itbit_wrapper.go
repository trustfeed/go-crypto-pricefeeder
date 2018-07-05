package itbit

import (
	"errors"
	"log"
	"strconv"
	"sync"

	"github.com/trustfeed/go-crypto-pricefeeder/currency/pair"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/orderbook"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/ticker"
)

// Start starts the ItBit go routine
func (i *ItBit) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		i.Run()
		wg.Done()
	}()
}

// Run implements the ItBit wrapper
func (i *ItBit) Run() {
	if i.Verbose {
		log.Printf("%s polling delay: %ds.\n", i.GetName(), i.RESTPollingDelay)
		log.Printf("%s %d currencies enabled: %s.\n", i.GetName(), len(i.EnabledPairs), i.EnabledPairs)
	}
}

// UpdateTicker updates and returns the ticker for a currency pair
func (i *ItBit) UpdateTicker(p pair.CurrencyPair, assetType string) (ticker.Price, error) {
	var tickerPrice ticker.Price
	tick, err := i.GetTicker(exchange.FormatExchangeCurrency(i.Name,
		p).String())
	if err != nil {
		return tickerPrice, err
	}

	tickerPrice.Pair = p
	tickerPrice.Ask = tick.Ask
	tickerPrice.Bid = tick.Bid
	tickerPrice.Last = tick.LastPrice
	tickerPrice.High = tick.High24h
	tickerPrice.Low = tick.Low24h
	tickerPrice.Volume = tick.Volume24h
	ticker.ProcessTicker(i.GetName(), p, tickerPrice, assetType)
	return ticker.GetTicker(i.Name, p, assetType)
}

// GetTickerPrice returns the ticker for a currency pair
func (i *ItBit) GetTickerPrice(p pair.CurrencyPair, assetType string) (ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(i.GetName(), p, assetType)
	if err != nil {
		return i.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// GetOrderbookEx returns orderbook base on the currency pair
func (i *ItBit) GetOrderbookEx(p pair.CurrencyPair, assetType string) (orderbook.Base, error) {
	ob, err := orderbook.GetOrderbook(i.GetName(), p, assetType)
	if err != nil {
		return i.UpdateOrderbook(p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (i *ItBit) UpdateOrderbook(p pair.CurrencyPair, assetType string) (orderbook.Base, error) {
	var orderBook orderbook.Base
	orderbookNew, err := i.GetOrderbook(exchange.FormatExchangeCurrency(i.Name,
		p).String())
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Bids {
		data := orderbookNew.Bids[x]
		price, err := strconv.ParseFloat(data[0], 64)
		if err != nil {
			log.Println(err)
		}
		amount, err := strconv.ParseFloat(data[1], 64)
		if err != nil {
			log.Println(err)
		}
		orderBook.Bids = append(orderBook.Bids, orderbook.Item{Amount: amount, Price: price})
	}

	for x := range orderbookNew.Asks {
		data := orderbookNew.Asks[x]
		price, err := strconv.ParseFloat(data[0], 64)
		if err != nil {
			log.Println(err)
		}
		amount, err := strconv.ParseFloat(data[1], 64)
		if err != nil {
			log.Println(err)
		}
		orderBook.Asks = append(orderBook.Asks, orderbook.Item{Amount: amount, Price: price})
	}

	orderbook.ProcessOrderbook(i.GetName(), p, orderBook, assetType)
	return orderbook.GetOrderbook(i.Name, p, assetType)
}

// GetExchangeAccountInfo retrieves balances for all enabled currencies for the
//ItBit exchange - to-do
func (i *ItBit) GetExchangeAccountInfo() (exchange.AccountInfo, error) {
	var response exchange.AccountInfo
	response.ExchangeName = i.GetName()
	return response, nil
}

// GetExchangeHistory returns historic trade data since exchange opening.
func (i *ItBit) GetExchangeHistory(p pair.CurrencyPair, assetType string) ([]exchange.TradeHistory, error) {
	var resp []exchange.TradeHistory

	return resp, errors.New("trade history not yet implemented")
}

// SubmitExchangeOrder submits a new order
func (i *ItBit) SubmitExchangeOrder(p pair.CurrencyPair, side string, orderType int, amount, price float64) (int64, error) {
	return 0, errors.New("not yet implemented")
}

// ModifyExchangeOrder will allow of changing orderbook placement and limit to
// market conversion
func (i *ItBit) ModifyExchangeOrder(p pair.CurrencyPair, orderID, action int64) (int64, error) {
	return 0, errors.New("not yet implemented")
}

// CancelExchangeOrder cancels an order by its corresponding ID number
func (i *ItBit) CancelExchangeOrder(p pair.CurrencyPair, orderID int64) (int64, error) {
	return 0, errors.New("not yet implemented")
}

// CancelAllExchangeOrders cancels all orders associated with a currency pair
func (i *ItBit) CancelAllExchangeOrders(p pair.CurrencyPair) error {
	return errors.New("not yet implemented")
}

// GetExchangeOrderInfo returns information on a current open order
func (i *ItBit) GetExchangeOrderInfo(orderID int64) (float64, error) {
	return 0, errors.New("not yet implemented")
}

// GetExchangeDepositAddress returns a deposit address for a specified currency
func (i *ItBit) GetExchangeDepositAddress(p pair.CurrencyPair) (string, error) {
	return "", errors.New("not yet implemented")
}

// WithdrawExchangeFunds returns a withdrawal ID when a withdrawal is submitted
func (i *ItBit) WithdrawExchangeFunds(address string, p pair.CurrencyPair, amount float64) (string, error) {
	return "", errors.New("not yet implemented")
}
