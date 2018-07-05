package binance

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/trustfeed/go-crypto-pricefeeder/common"
	"github.com/trustfeed/go-crypto-pricefeeder/config"
	exchange "github.com/trustfeed/go-crypto-pricefeeder/exchanges"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/request"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/ticker"
)

// Binance is the overarching type across the Bithumb package
type Binance struct {
	exchange.Base

	// valid string list that a required by the exchange
	validLimits    []string
	validIntervals []string
}

const (
	apiURL = "https://api.binance.com"

	// Public endpoints
	exchangeInfo     = "/api/v1/exchangeInfo"
	orderBookDepth   = "/api/v1/depth"
	recentTrades     = "/api/v1/trades"
	historicalTrades = "/api/v1/historicalTrades"
	aggregatedTrades = "/api/v1/aggTrades"
	candleStick      = "/api/v1/klines"
	priceChange      = "/api/v1/ticker/24hr"
	symbolPrice      = "/api/v3/ticker/price"
	bestPrice        = "/api/v3/ticker/bookTicker"
	accountInfo      = "/api/v3/account"

	// Authenticated endpoints
	newOrderTest = "/api/v3/order/test"
	newOrder     = "/api/v3/order"
	queryOrder   = "/api/v3/order"

	// binance authenticated and unauthenticated limit rates
	// to-do
	binanceAuthRate   = 0
	binanceUnauthRate = 0
)

// SetDefaults sets the basic defaults for Binance
func (b *Binance) SetDefaults() {
	b.Name = "Binance"
	b.Enabled = false
	b.Verbose = false
	b.Websocket = false
	b.RESTPollingDelay = 10
	b.RequestCurrencyPairFormat.Delimiter = ""
	b.RequestCurrencyPairFormat.Uppercase = true
	b.ConfigCurrencyPairFormat.Delimiter = "-"
	b.ConfigCurrencyPairFormat.Uppercase = true
	b.AssetTypes = []string{ticker.Spot}
	b.SupportsAutoPairUpdating = true
	b.SupportsRESTTickerBatching = true
	b.SetValues()
	b.Requester = request.New(b.Name, request.NewRateLimit(time.Second, binanceAuthRate), request.NewRateLimit(time.Second, binanceUnauthRate), common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))
}

// Setup takes in the supplied exchange configuration details and sets params
func (b *Binance) Setup(exch config.ExchangeConfig) {
	if !exch.Enabled {
		b.SetEnabled(false)
	} else {
		b.Enabled = true
		b.AuthenticatedAPISupport = exch.AuthenticatedAPISupport
		b.SetAPIKeys(exch.APIKey, exch.APISecret, "", false)
		b.SetHTTPClientTimeout(exch.HTTPTimeout)
		b.RESTPollingDelay = exch.RESTPollingDelay
		b.Verbose = exch.Verbose
		b.Websocket = exch.Websocket
		b.BaseCurrencies = common.SplitStrings(exch.BaseCurrencies, ",")
		b.AvailablePairs = common.SplitStrings(exch.AvailablePairs, ",")
		b.EnabledPairs = common.SplitStrings(exch.EnabledPairs, ",")
		err := b.SetCurrencyPairFormat()
		if err != nil {
			log.Fatal(err)
		}
		err = b.SetAssetTypes()
		if err != nil {
			log.Fatal(err)
		}
		err = b.SetAutoPairDefaults()
		if err != nil {
			log.Fatal(err)
		}
	}
}

// GetExchangeValidCurrencyPairs returns the full pair list from the exchange
// at the moment do not integrate with config currency pairs automatically
func (b *Binance) GetExchangeValidCurrencyPairs() ([]string, error) {
	var validCurrencyPairs []string

	info, err := b.GetExchangeInfo()
	if err != nil {
		return nil, err
	}

	for _, symbol := range info.Symbols {
		if symbol.Status == "TRADING" {
			validCurrencyPairs = append(validCurrencyPairs, symbol.BaseAsset+"-"+symbol.QuoteAsset)
		}
	}
	return validCurrencyPairs, nil
}

// GetExchangeInfo returns exchange information. Check binance_types for more
// information
func (b *Binance) GetExchangeInfo() (ExchangeInfo, error) {
	var resp ExchangeInfo
	path := apiURL + exchangeInfo

	return resp, b.SendHTTPRequest(path, &resp)
}

// GetOrderBook returns full orderbook information
//
// symbol: string of currency pair
// limit: returned limit amount
func (b *Binance) GetOrderBook(symbol string, limit int64) (OrderBook, error) {
	orderbook, resp := OrderBook{}, OrderBookData{}

	if err := b.CheckLimit(limit); err != nil {
		return orderbook, err
	}
	if err := b.CheckSymbol(symbol); err != nil {
		return orderbook, err
	}

	params := url.Values{}
	params.Set("symbol", common.StringToUpper(symbol))
	params.Set("limit", strconv.FormatInt(limit, 10))

	path := fmt.Sprintf("%s%s?%s", apiURL, orderBookDepth, params.Encode())

	if err := b.SendHTTPRequest(path, &resp); err != nil {
		return orderbook, err
	}

	for _, asks := range resp.Asks {
		var ASK struct {
			Price    float64
			Quantity float64
		}
		for i, ask := range asks.([]interface{}) {
			switch i {
			case 0:
				ASK.Price, _ = strconv.ParseFloat(ask.(string), 64)
			case 1:
				ASK.Quantity, _ = strconv.ParseFloat(ask.(string), 64)
				orderbook.Asks = append(orderbook.Asks, ASK)
				break
			}
		}
	}

	for _, bids := range resp.Bids {
		var BID struct {
			Price    float64
			Quantity float64
		}
		for i, bid := range bids.([]interface{}) {
			switch i {
			case 0:
				BID.Price, _ = strconv.ParseFloat(bid.(string), 64)
			case 1:
				BID.Quantity, _ = strconv.ParseFloat(bid.(string), 64)
				orderbook.Bids = append(orderbook.Bids, BID)
				break
			}
		}
	}
	return orderbook, nil
}

// GetRecentTrades returns recent trade activity
//
// symbol: string of currency pair
// limit: returned limit amount WARNING: MAX 500!
func (b *Binance) GetRecentTrades(symbol string, limit int64) ([]RecentTrade, error) {
	resp := []RecentTrade{}

	if err := b.CheckLimit(limit); err != nil {
		return resp, err
	}
	if err := b.CheckSymbol(symbol); err != nil {
		return resp, err
	}

	params := url.Values{}
	params.Set("symbol", common.StringToUpper(symbol))
	params.Set("limit", strconv.FormatInt(limit, 10))

	path := fmt.Sprintf("%s%s?%s", apiURL, recentTrades, params.Encode())

	return resp, b.SendHTTPRequest(path, &resp)
}

// GetHistoricalTrades returns historical trade activity
//
// symbol: string of currency pair
// limit: returned limit amount WARNING: MAX 500! (NOT REQUIRED)
// fromID:
func (b *Binance) GetHistoricalTrades(symbol string, limit, fromID int64) ([]HistoricalTrade, error) {
	resp := []HistoricalTrade{}

	if err := b.CheckLimit(limit); err != nil {
		return resp, err
	}
	if err := b.CheckSymbol(symbol); err != nil {
		return resp, err
	}

	params := url.Values{}
	params.Set("symbol", common.StringToUpper(symbol))
	params.Set("limit", strconv.FormatInt(limit, 10))
	params.Set("fromid", strconv.FormatInt(fromID, 10))

	path := fmt.Sprintf("%s%s?%s", apiURL, historicalTrades, params.Encode())

	return resp, b.SendHTTPRequest(path, &resp)
}

// GetAggregatedTrades returns aggregated trade activity
//
// symbol: string of currency pair
// limit: returned limit amount WARNING: MAX 500!
func (b *Binance) GetAggregatedTrades(symbol string, limit int64) ([]AggregatedTrade, error) {
	resp := []AggregatedTrade{}

	if err := b.CheckLimit(limit); err != nil {
		return resp, err
	}
	if err := b.CheckSymbol(symbol); err != nil {
		return resp, err
	}

	params := url.Values{}
	params.Set("symbol", common.StringToUpper(symbol))
	params.Set("limit", strconv.FormatInt(limit, 10))

	path := fmt.Sprintf("%s%s?%s", apiURL, aggregatedTrades, params.Encode())

	return resp, b.SendHTTPRequest(path, &resp)
}

// GetCandleStickData returns candle stick data
//
// symbol:
// limit:
// interval
func (b *Binance) GetCandleStickData(symbol, interval string, limit int64) ([]CandleStick, error) {
	var resp interface{}
	var kline []CandleStick

	if err := b.CheckLimit(limit); err != nil {
		return kline, err
	}
	if err := b.CheckSymbol(symbol); err != nil {
		return kline, err
	}
	if err := b.CheckIntervals(interval); err != nil {
		return kline, err
	}

	params := url.Values{}
	params.Set("symbol", common.StringToUpper(symbol))
	params.Set("limit", strconv.FormatInt(limit, 10))
	params.Set("interval", interval)

	path := fmt.Sprintf("%s%s?%s", apiURL, candleStick, params.Encode())

	if err := b.SendHTTPRequest(path, &resp); err != nil {
		return kline, err
	}

	for _, responseData := range resp.([]interface{}) {
		var candle CandleStick
		for i, individualData := range responseData.([]interface{}) {
			switch i {
			case 0:
				candle.OpenTime = individualData.(float64)
			case 1:
				candle.Open, _ = strconv.ParseFloat(individualData.(string), 64)
			case 2:
				candle.High, _ = strconv.ParseFloat(individualData.(string), 64)
			case 3:
				candle.Low, _ = strconv.ParseFloat(individualData.(string), 64)
			case 4:
				candle.Close, _ = strconv.ParseFloat(individualData.(string), 64)
			case 5:
				candle.Volume, _ = strconv.ParseFloat(individualData.(string), 64)
			case 6:
				candle.CloseTime = individualData.(float64)
			case 7:
				candle.QuoteAssetVolume, _ = strconv.ParseFloat(individualData.(string), 64)
			case 8:
				candle.TradeCount = individualData.(float64)
			case 9:
				candle.TakerBuyAssetVolume, _ = strconv.ParseFloat(individualData.(string), 64)
			case 10:
				candle.TakerBuyQuoteAssetVolume, _ = strconv.ParseFloat(individualData.(string), 64)
			}
		}
		kline = append(kline, candle)
	}
	return kline, nil
}

// GetPriceChangeStats returns price change statistics for the last 24 hours
//
// symbol: string of currency pair
func (b *Binance) GetPriceChangeStats(symbol string) (PriceChangeStats, error) {
	resp := PriceChangeStats{}

	if err := b.CheckSymbol(symbol); err != nil {
		return resp, err
	}

	params := url.Values{}
	params.Set("symbol", common.StringToUpper(symbol))

	path := fmt.Sprintf("%s%s?%s", apiURL, priceChange, params.Encode())

	return resp, b.SendHTTPRequest(path, &resp)
}

// GetTickers returns the ticker data for the last 24 hrs
func (b *Binance) GetTickers() ([]PriceChangeStats, error) {
	var resp []PriceChangeStats
	path := fmt.Sprintf("%s%s", apiURL, priceChange)
	return resp, b.SendHTTPRequest(path, &resp)
}

// GetLatestSpotPrice returns latest spot price of symbol
//
// symbol: string of currency pair
func (b *Binance) GetLatestSpotPrice(symbol string) (SymbolPrice, error) {
	resp := SymbolPrice{}

	if err := b.CheckSymbol(symbol); err != nil {
		return resp, err
	}

	params := url.Values{}
	params.Set("symbol", common.StringToUpper(symbol))

	path := fmt.Sprintf("%s%s?%s", apiURL, symbolPrice, params.Encode())

	return resp, b.SendHTTPRequest(path, &resp)
}

// GetBestPrice returns the latest best price for symbol
//
// symbol: string of currency pair
func (b *Binance) GetBestPrice(symbol string) (BestPrice, error) {
	resp := BestPrice{}

	if err := b.CheckSymbol(symbol); err != nil {
		return resp, err
	}

	params := url.Values{}
	params.Set("symbol", common.StringToUpper(symbol))

	path := fmt.Sprintf("%s%s?%s", apiURL, bestPrice, params.Encode())

	return resp, b.SendHTTPRequest(path, &resp)
}

// NewOrderTest sends a new order
func (b *Binance) NewOrderTest() (interface{}, error) {
	var resp interface{}

	path := fmt.Sprintf("%s%s", apiURL, newOrderTest)

	params := url.Values{}
	params.Set("symbol", "BTCUSDT")
	params.Set("side", "BUY")
	params.Set("type", "MARKET")
	params.Set("quantity", "0.1")

	return resp, b.SendAuthHTTPRequest("POST", path, params, &resp)
}

// NewOrder sends a new order to Binance
func (b *Binance) NewOrder(o NewOrderRequest) (NewOrderResponse, error) {
	var resp NewOrderResponse

	path := fmt.Sprintf("%s%s", apiURL, newOrderTest)

	params := url.Values{}
	params.Set("symbol", o.Symbol)
	params.Set("side", o.Side)
	params.Set("type", o.TradeType)
	params.Set("timeInForce", o.TimeInForce)
	params.Set("quantity", strconv.FormatFloat(o.Quantity, 'f', -1, 64))
	params.Set("price", strconv.FormatFloat(o.Price, 'f', -1, 64))
	params.Set("newClientOrderID", o.NewClientOrderID)
	params.Set("stopPrice", strconv.FormatFloat(o.StopPrice, 'f', -1, 64))
	params.Set("icebergQty", strconv.FormatFloat(o.IcebergQty, 'f', -1, 64))
	params.Set("newOrderRespType", o.NewOrderRespType)

	if err := b.SendAuthHTTPRequest("POST", path, params, &resp); err != nil {
		return resp, err
	}

	if resp.Code != 0 {
		return resp, errors.New(resp.Msg)
	}
	return resp, nil
}

// QueryOrder returns information on a past order
func (b *Binance) QueryOrder(symbol, origClientOrderID string, orderID int64) (QueryOrderData, error) {
	var resp QueryOrderData

	path := fmt.Sprintf("%s%s", apiURL, queryOrder)

	params := url.Values{}
	params.Set("symbol", common.StringToUpper(symbol))
	params.Set("origClientOrderId", origClientOrderID)
	params.Set("orderId", strconv.FormatInt(orderID, 10))

	if err := b.SendAuthHTTPRequest("GET", path, params, &resp); err != nil {
		return resp, err
	}

	if resp.Code != 0 {
		return resp, errors.New(resp.Msg)
	}
	return resp, nil
}

// GetAccount returns binance user accounts
func (b *Binance) GetAccount() (*Account, error) {
	type respone struct {
		Response
		Account
	}

	var resp respone

	path := fmt.Sprintf("%s%s", apiURL, accountInfo)
	params := url.Values{}

	if err := b.SendAuthHTTPRequest("GET", path, params, &resp); err != nil {
		return &resp.Account, err
	}

	if resp.Code != 0 {
		return &resp.Account, errors.New(resp.Msg)
	}

	return &resp.Account, nil
}

// SendHTTPRequest sends an unauthenticated request
func (b *Binance) SendHTTPRequest(path string, result interface{}) error {
	return b.SendPayload("GET", path, nil, nil, result, false, b.Verbose)
}

// SendAuthHTTPRequest sends an authenticated HTTP request
func (b *Binance) SendAuthHTTPRequest(method, path string, params url.Values, result interface{}) error {
	if !b.AuthenticatedAPISupport {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet, b.Name)
	}

	if params == nil {
		params = url.Values{}
	}
	params.Set("recvWindow", strconv.FormatInt(5000, 10))
	params.Set("timestamp", strconv.FormatInt(time.Now().Unix()*1000, 10))

	signature := params.Encode()
	hmacSigned := common.GetHMAC(common.HashSHA256, []byte(signature), []byte(b.APISecret))
	hmacSignedStr := common.HexEncodeToString(hmacSigned)
	params.Set("signature", hmacSignedStr)

	headers := make(map[string]string)
	headers["X-MBX-APIKEY"] = b.APIKey
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	if b.Verbose {
		log.Printf("sent path: \n%s\n", path)
	}
	path = common.EncodeURLValues(path, params)

	return b.SendPayload(method, path, headers, bytes.NewBufferString(""), result, true, b.Verbose)
}

// CheckLimit checks value against a variable list
func (b *Binance) CheckLimit(limit int64) error {
	if !common.StringDataCompare(b.validLimits, strconv.FormatInt(limit, 10)) {
		return errors.New("Incorrect limit values - valid values are 5, 10, 20, 50, 100, 500, 1000")
	}
	return nil
}

// CheckSymbol checks value against a variable list
func (b *Binance) CheckSymbol(symbol string) error {
	enPairs := b.GetEnabledCurrencies()
	for x := range enPairs {
		if exchange.FormatExchangeCurrency(b.Name, enPairs[x]).String() == symbol {
			return nil
		}
	}
	return errors.New("Incorrect symbol values - please check available pairs in configuration")
}

// CheckIntervals checks value against a variable list
func (b *Binance) CheckIntervals(interval string) error {
	if !common.StringDataCompare(b.validIntervals, interval) {
		return errors.New(`Incorrect interval values - valid values are "1m","3m","5m","15m","30m","1h","2h","4h","6h","8h","12h","1d","3d","1w","1M"`)
	}
	return nil
}

// SetValues sets the default valid values
func (b *Binance) SetValues() {
	b.validLimits = []string{"5", "10", "20", "50", "100", "500", "1000"}
	b.validIntervals = []string{"1m", "3m", "5m", "15m", "30m", "1h", "2h", "4h", "6h", "8h", "12h", "1d", "3d", "1w", "1M"}
}
