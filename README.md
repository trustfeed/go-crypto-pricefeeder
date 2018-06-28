# go-crypto-pricefeeder
A cryptocurrency trading bot supporting multiple exchanges written in Golang.

## Exchange Support Table

| Exchange | REST API | Streaming API | FIX API |
|----------|------|-----------|-----|
| Alphapoint | Yes  | Yes        | NA  |
| ANXPRO | Yes  | No        | NA  |
| Binance| Yes  | No        | NA  |
| Bitfinex | Yes  | Yes        | NA  |
| Bitflyer | Yes  | No      | NA  |
| Bithumb | Yes  | NA       | NA  |
| Bitstamp | Yes  | Yes       | No  |
| Bittrex | Yes | No | NA |
| BTCC | Yes  | Yes     | No  |
| BTCMarkets | Yes | No       | NA  |
| COINUT | Yes | No | NA |
| Exmo | Yes | NA | NA |
| GDAX(Coinbase) | Yes | Yes | No|
| Gemini | Yes | No | No |
| HitBTC | Yes | Yes | No |
| Huobi.Pro | Yes | No |No |
| ItBit | Yes | NA | No |
| Kraken | Yes | NA | NA |
| LakeBTC | Yes | No | NA |
| Liqui | Yes | No | NA |
| LocalBitcoins | Yes | NA | NA |
| OKCoin China | Yes | Yes | No |
| OKCoin International | Yes | Yes | No |
| OKEX | Yes | No | No |
| Poloniex | Yes | Yes | NA |
| WEX     | Yes  | NA        | NA  |
| Yobit | Yes | NA | NA |

We are aiming to support the top 20 highest volume exchanges based off the [CoinMarketCap exchange data](https://coinmarketcap.com/exchanges/volume/24-hour/).

** NA means not applicable as the Exchange does not support the feature.

## Current Features

+ Support for all Exchange fiat and digital currencies, with the ability to individually toggle them on/off.
+ AES encrypted config file.
+ REST API support for all exchanges.
+ Websocket support for applicable exchanges.
+ Ability to turn off/on certain exchanges.
+ Ability to adjust manual polling timer for exchanges.
+ SMS notification support via SMS Gateway.
+ Packages for handling currency pairs, ticker/orderbook fetching and currency conversion.
+ Portfolio management tool; fetches balances from supported exchanges and allows for custom address tracking.
+ Basic event trigger system.
+ WebGUI.

## Compiling instructions

Download and install Go from [Go Downloads](https://golang.org/dl/) for your
platform.

### Linux/OSX

We use the `dep` tool provided by Golang for managing dependencies. As it is not officially part
of the go tools package suite, you will need to manually install it if you have not already.

On MacOS you can install or upgrade to the latest released version with Homebrew:

```sh
brew install dep
brew upgrade dep
```

On linux or MacOS, you can also install it via `go get`:

```sh
go get -u github.com/golang/dep/cmd/dep
```

After `dep` is installed, please follow the instructions below:

```bash
go get github.com/thrasher-/gocryptotrader
cd $GOPATH/src/github.com/thrasher-/gocryptotrader
make get
make install
cp $GOPATH/src/github.com/thrasher-/gocryptotrader/config_example.json $GOPATH/bin/config.json
```

+ Make any neccessary changes to the `config.json` file.
+ Run the `gocryptotrader` binary file inside your GOPATH bin folder.