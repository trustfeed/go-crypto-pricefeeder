package wex

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/trustfeed/go-crypto-pricefeeder/common"
	"github.com/trustfeed/go-crypto-pricefeeder/config"
	exchange "github.com/trustfeed/go-crypto-pricefeeder/exchanges"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/request"
	"github.com/trustfeed/go-crypto-pricefeeder/exchanges/ticker"
)

const (
	wexAPIPublicURL       = "https://wex.nz/api"
	wexAPIPrivateURL      = "https://wex.nz/tapi"
	wexAPIPublicVersion   = "3"
	wexAPIPrivateVersion  = "1"
	wexInfo               = "info"
	wexTicker             = "ticker"
	wexDepth              = "depth"
	wexTrades             = "trades"
	wexAccountInfo        = "getInfo"
	wexTrade              = "Trade"
	wexActiveOrders       = "ActiveOrders"
	wexOrderInfo          = "OrderInfo"
	wexCancelOrder        = "CancelOrder"
	wexTradeHistory       = "TradeHistory"
	wexTransactionHistory = "TransHistory"
	wexWithdrawCoin       = "WithdrawCoin"
	wexCoinDepositAddress = "CoinDepositAddress"
	wexCreateCoupon       = "CreateCoupon"
	wexRedeemCoupon       = "RedeemCoupon"

	wexAuthRate   = 0
	wexUnauthRate = 0
)

// WEX is the overarching type across the wex package
type WEX struct {
	exchange.Base
	Ticker map[string]Ticker
}

// SetDefaults sets current default value for WEX
func (w *WEX) SetDefaults() {
	w.Name = "WEX"
	w.Enabled = false
	w.Fee = 0.2
	w.Verbose = false
	w.Websocket = false
	w.RESTPollingDelay = 10
	w.Ticker = make(map[string]Ticker)
	w.RequestCurrencyPairFormat.Delimiter = "_"
	w.RequestCurrencyPairFormat.Uppercase = false
	w.RequestCurrencyPairFormat.Separator = "-"
	w.ConfigCurrencyPairFormat.Delimiter = "_"
	w.ConfigCurrencyPairFormat.Uppercase = true
	w.AssetTypes = []string{ticker.Spot}
	w.SupportsAutoPairUpdating = true
	w.SupportsRESTTickerBatching = true
	w.Requester = request.New(w.Name, request.NewRateLimit(time.Second, wexAuthRate), request.NewRateLimit(time.Second, wexUnauthRate), common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))
}

// Setup sets exchange configuration parameters for WEX
func (w *WEX) Setup(exch config.ExchangeConfig) {
	if !exch.Enabled {
		w.SetEnabled(false)
	} else {
		w.Enabled = true
		w.AuthenticatedAPISupport = exch.AuthenticatedAPISupport
		w.SetAPIKeys(exch.APIKey, exch.APISecret, "", false)
		w.SetHTTPClientTimeout(exch.HTTPTimeout)
		w.RESTPollingDelay = exch.RESTPollingDelay
		w.Verbose = exch.Verbose
		w.Websocket = exch.Websocket
		w.BaseCurrencies = common.SplitStrings(exch.BaseCurrencies, ",")
		w.AvailablePairs = common.SplitStrings(exch.AvailablePairs, ",")
		w.EnabledPairs = common.SplitStrings(exch.EnabledPairs, ",")
		err := w.SetCurrencyPairFormat()
		if err != nil {
			log.Fatal(err)
		}
		err = w.SetAssetTypes()
		if err != nil {
			log.Fatal(err)
		}
		err = w.SetAutoPairDefaults()
		if err != nil {
			log.Fatal(err)
		}
	}
}

// GetTradablePairs returns a list of available pairs from the exchange
func (w *WEX) GetTradablePairs() ([]string, error) {
	result, err := w.GetInfo()
	if err != nil {
		return nil, err
	}

	var currencies []string
	for x := range result.Pairs {
		currencies = append(currencies, common.StringToUpper(x))
	}

	return currencies, nil
}

// GetFee returns the exchange fee
func (w *WEX) GetFee() float64 {
	return w.Fee
}

// GetInfo returns the WEX info
func (w *WEX) GetInfo() (Info, error) {
	resp := Info{}
	req := fmt.Sprintf("%s/%s/%s/", wexAPIPublicURL, wexAPIPublicVersion, wexInfo)

	return resp, w.SendHTTPRequest(req, &resp)
}

// GetTicker returns a ticker for a specific currency
func (w *WEX) GetTicker(symbol string) (map[string]Ticker, error) {
	type Response struct {
		Data map[string]Ticker
	}

	response := Response{}
	req := fmt.Sprintf("%s/%s/%s/%s", wexAPIPublicURL, wexAPIPublicVersion, wexTicker, symbol)

	return response.Data, w.SendHTTPRequest(req, &response.Data)
}

// GetDepth returns the depth for a specific currency
func (w *WEX) GetDepth(symbol string) (Orderbook, error) {
	type Response struct {
		Data map[string]Orderbook
	}

	response := Response{}
	req := fmt.Sprintf("%s/%s/%s/%s", wexAPIPublicURL, wexAPIPublicVersion, wexDepth, symbol)

	return response.Data[symbol], w.SendHTTPRequest(req, &response.Data)
}

// GetTrades returns the trades for a specific currency
func (w *WEX) GetTrades(symbol string) ([]Trades, error) {
	type Response struct {
		Data map[string][]Trades
	}

	response := Response{}
	req := fmt.Sprintf("%s/%s/%s/%s", wexAPIPublicURL, wexAPIPublicVersion, wexTrades, symbol)

	return response.Data[symbol], w.SendHTTPRequest(req, &response.Data)
}

// GetAccountInfo returns a users account info
func (w *WEX) GetAccountInfo() (AccountInfo, error) {
	var result AccountInfo

	err := w.SendAuthenticatedHTTPRequest(wexAccountInfo, url.Values{}, &result)
	if err != nil {
		return result, err
	}

	if result.Error != "" {
		return result, errors.New(result.Error)
	}
	return result, nil
}

// GetActiveOrders returns the active orders for a specific currency
func (w *WEX) GetActiveOrders(pair string) (map[string]ActiveOrders, error) {
	req := url.Values{}
	req.Add("pair", pair)

	var result map[string]ActiveOrders

	return result, w.SendAuthenticatedHTTPRequest(wexActiveOrders, req, &result)
}

// GetOrderInfo returns the order info for a specific order ID
func (w *WEX) GetOrderInfo(OrderID int64) (map[string]OrderInfo, error) {
	req := url.Values{}
	req.Add("order_id", strconv.FormatInt(OrderID, 10))

	var result map[string]OrderInfo

	return result, w.SendAuthenticatedHTTPRequest(wexOrderInfo, req, &result)
}

// CancelOrder cancels an order for a specific order ID
func (w *WEX) CancelOrder(OrderID int64) (bool, error) {
	req := url.Values{}
	req.Add("order_id", strconv.FormatInt(OrderID, 10))

	var result CancelOrder

	err := w.SendAuthenticatedHTTPRequest(wexCancelOrder, req, &result)
	if err != nil {
		return false, err
	}

	if result.Error != "" {
		return false, errors.New(result.Error)
	}
	return true, nil
}

// Trade places an order and returns the order ID if successful or an error
func (w *WEX) Trade(pair, orderType string, amount, price float64) (int64, error) {
	req := url.Values{}
	req.Add("pair", pair)
	req.Add("type", orderType)
	req.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	req.Add("rate", strconv.FormatFloat(price, 'f', -1, 64))

	var result Trade

	err := w.SendAuthenticatedHTTPRequest(wexTrade, req, &result)
	if err != nil {
		return 0, err
	}

	if result.Error != "" {
		return 0, errors.New(result.Error)
	}
	return int64(result.OrderID), nil
}

// GetTransactionHistory returns the transaction history
func (w *WEX) GetTransactionHistory(TIDFrom, Count, TIDEnd int64, order, since, end string) (map[string]TransHistory, error) {
	req := url.Values{}
	req.Add("from", strconv.FormatInt(TIDFrom, 10))
	req.Add("count", strconv.FormatInt(Count, 10))
	req.Add("from_id", strconv.FormatInt(TIDFrom, 10))
	req.Add("end_id", strconv.FormatInt(TIDEnd, 10))
	req.Add("order", order)
	req.Add("since", since)
	req.Add("end", end)

	var result map[string]TransHistory

	return result,
		w.SendAuthenticatedHTTPRequest(wexTransactionHistory, req, &result)
}

// GetTradeHistory returns the trade history
func (w *WEX) GetTradeHistory(TIDFrom, Count, TIDEnd int64, order, since, end, pair string) (map[string]TradeHistory, error) {
	req := url.Values{}
	req.Add("from", strconv.FormatInt(TIDFrom, 10))
	req.Add("count", strconv.FormatInt(Count, 10))
	req.Add("from_id", strconv.FormatInt(TIDFrom, 10))
	req.Add("end_id", strconv.FormatInt(TIDEnd, 10))
	req.Add("order", order)
	req.Add("since", since)
	req.Add("end", end)
	req.Add("pair", pair)

	var result map[string]TradeHistory

	return result, w.SendAuthenticatedHTTPRequest(wexTradeHistory, req, &result)
}

// WithdrawCoins withdraws coins for a specific coin
func (w *WEX) WithdrawCoins(coin string, amount float64, address string) (WithdrawCoins, error) {
	req := url.Values{}
	req.Add("coinName", coin)
	req.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	req.Add("address", address)

	var result WithdrawCoins

	err := w.SendAuthenticatedHTTPRequest(wexWithdrawCoin, req, &result)
	if err != nil {
		return result, err
	}

	if result.Error != "" {
		return result, errors.New(result.Error)
	}
	return result, nil
}

// CoinDepositAddress returns the deposit address for a specific currency
func (w *WEX) CoinDepositAddress(coin string) (string, error) {
	req := url.Values{}
	req.Add("coinName", coin)

	var result CoinDepositAddress

	err := w.SendAuthenticatedHTTPRequest(wexCoinDepositAddress, req, &result)
	if err != nil {
		return result.Address, err
	}
	if result.Error != "" {
		return result.Address, errors.New(result.Error)
	}
	return result.Address, nil
}

// CreateCoupon creates an exchange coupon for a sepcific currency
func (w *WEX) CreateCoupon(currency string, amount float64) (CreateCoupon, error) {
	req := url.Values{}
	req.Add("currency", currency)
	req.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))

	var result CreateCoupon

	err := w.SendAuthenticatedHTTPRequest(wexCreateCoupon, req, &result)
	if err != nil {
		return result, err
	}
	if result.Error != "" {
		return result, errors.New(result.Error)
	}
	return result, nil
}

// RedeemCoupon redeems an exchange coupon
func (w *WEX) RedeemCoupon(coupon string) (RedeemCoupon, error) {
	req := url.Values{}
	req.Add("coupon", coupon)

	var result RedeemCoupon

	err := w.SendAuthenticatedHTTPRequest(wexRedeemCoupon, req, &result)
	if err != nil {
		return result, err
	}
	if result.Error != "" {
		return result, errors.New(result.Error)
	}
	return result, nil
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (w *WEX) SendHTTPRequest(path string, result interface{}) error {
	return w.SendPayload("GET", path, nil, nil, result, false, w.Verbose)
}

// SendAuthenticatedHTTPRequest sends an authenticated HTTP request to WEX
func (w *WEX) SendAuthenticatedHTTPRequest(method string, values url.Values, result interface{}) (err error) {
	if !w.AuthenticatedAPISupport {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet, w.Name)
	}

	if w.Nonce.Get() == 0 {
		w.Nonce.Set(time.Now().Unix())
	} else {
		w.Nonce.Inc()
	}
	values.Set("nonce", w.Nonce.String())
	values.Set("method", method)

	encoded := values.Encode()
	hmac := common.GetHMAC(common.HashSHA512, []byte(encoded), []byte(w.APISecret))

	if w.Verbose {
		log.Printf("Sending POST request to %s calling method %s with params %s\n",
			wexAPIPrivateURL,
			method,
			encoded)
	}

	headers := make(map[string]string)
	headers["Key"] = w.APIKey
	headers["Sign"] = common.HexEncodeToString(hmac)
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	return w.SendPayload("POST", wexAPIPrivateURL, headers, strings.NewReader(encoded), result, true, w.Verbose)
}
