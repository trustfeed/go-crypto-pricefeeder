package telegram

import (
	"testing"

	"github.com/trustfeed/go-crypto-pricefeeder/communications/base"
	"github.com/trustfeed/go-crypto-pricefeeder/config"
)

var T Telegram

func TestSetup(t *testing.T) {
	cfg := config.GetConfig()
	cfg.LoadConfig("../../testdata/configtest.json")
	T.Setup(cfg.GetCommunicationsConfig())
}

func TestConnect(t *testing.T) {
	err := T.Connect()
	if err == nil {
		t.Error("test failed - telegram Connect() error", err)
	}
}

func PushEvent(t *testing.T) {
	err := T.PushEvent(base.Event{})
	if err != nil {
		t.Error("test failed - telegram PushEvent() error", err)
	}
}
