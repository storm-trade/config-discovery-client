package main

import (
	"fmt"
	"github.com/storm-trade/config-discovery-client/client"
)

func main() {
	config, err := client.New("http://localhost:55824/config")
	if err != nil {
		panic(fmt.Errorf("can't initialize storm config: %w", err))
	}

	markets := config.GetMarketsByAssetName("LTC")

	for _, market := range markets {
		fmt.Printf("Market %s: address=%s, vaultAddress=%s, settlementToken=%s \n", market.Name, market.Address, market.VaultAddress, market.SettlementToken)
	}
}
