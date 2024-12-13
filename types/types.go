package types

type AssetSchedule struct {
	ScheduleTimeZone string `json:"scheduleTimeZone,omitempty"`
	Schedule         string `json:"schedule"`
	Holidays         string `json:"holidays"`
}

type AssetsSchedule struct {
	Schedules map[string]*AssetSchedule `json:"schedules"`
}

type Asset struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
	Type  string `json:"type"`
}

type CollateralAsset struct {
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
	AssetId  string `json:"assetId"`
}

type Market struct {
	Name            string   `json:"name"`
	Ticker          string   `json:"ticker"`
	Address         string   `json:"address"`
	VaultAddress    string   `json:"vaultAddress"`
	ImageLink       string   `json:"imageLink"`
	QuoteAsset      string   `json:"quoteAsset"`
	QuoteAssetId    string   `json:"quoteAssetId"`
	BaseAsset       string   `json:"baseAsset"`
	SettlementToken string   `json:"settlementToken"`
	Tags            []string `json:"tags"`
	Type            string   `json:"type"`
}

type Vault struct {
	Asset          CollateralAsset `json:"asset"`
	VaultAddress   string          `json:"vaultAddress"`
	QuoteAssetId   string          `json:"quoteAssetId"`
	LpJettonMaster string          `json:"lpJettonMaster"`
}

type AppConfig struct {
	ComposedAt       string            `json:"composedAt"`
	CollateralAssets []CollateralAsset `json:"assets"`
	OpenedMarkets    []Market          `json:"openedMarkets"`
	Vaults           []Vault           `json:"liquiditySources"`
}
