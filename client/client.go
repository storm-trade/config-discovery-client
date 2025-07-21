package client

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/storm-trade/config-discovery-client/request"
	"github.com/storm-trade/config-discovery-client/types"
	"golang.org/x/exp/maps"
	"time"
)

type ConfigDiscovery interface {
	ListenUpdates() error
	GetConfig() *types.AppConfig
	GetAssets() []*types.Asset
	GetSchedules() map[string]*types.AssetSchedule
	HasMarketByAddress(address string) bool
	GetMarketByAddress(address string) *types.Market
	HasPrelaunchMarketByAddress(address string) bool
	GetPrelaunchMarketByAddress(address string) *types.Market
	GetMarketsAddresses() []string
	GetMarketsByAssetName(name string) []types.Market
	HasVaultByAddress(address string) bool
	GetVaultByAddress(address string) *types.Vault
	HasVaultByLpJettonMasterAddress(address string) bool
	GetVaultByLpJettonMasterAddress(address string) *types.Vault
	HasAssetByIndex(index int) bool
	GetAssetByIndex(index int) *types.Asset
	HasAssetByName(name string) bool
	GetAssetByName(name string) *types.Asset
	HasCollateralAssetByName(name string) bool
	GetCollateralAssetByName(name string) *types.CollateralAsset
	HasVaultByCollateralAssetId(assetId string) bool
	GetVaultByCollateralAssetId(assetId string) *types.Vault
	HasVaultByCollateralAssetName(name string) bool
	GetVaultByCollateralAssetName(name string) *types.Vault
	UpdatesChannel() <-chan *types.AppConfig
}

type configDiscovery struct {
	cfgUri        string
	LastUpdatedAt *string
	Config        *types.AppConfig
	Assets        []*types.Asset
	AssetConfigs  []*types.AssetConfig
	Schedules     map[string]*types.AssetSchedule

	Updates chan *types.AppConfig
	// Maps
	VaultsMapByAddress               map[string]*types.Vault
	VaultsMapByCollateralAssetName   map[string]*types.Vault
	VaultsMapByCollateralAssetId     map[string]*types.Vault
	VaultsMapByLpJettonMasterAddress map[string]*types.Vault
	MarketsMapByAddress              map[string]*types.Market
	PrelaunchMarketsMapByAddress     map[string]*types.Market
	MarketsMapByBaseAssetName        map[string][]types.Market
	AssetsMapByName                  map[string]*types.Asset
	AssetsMapByIndex                 map[int]*types.Asset
	CollateralAssetsMapByName        map[string]*types.CollateralAsset
	AssetConfigsMapByName            map[string]*types.AssetConfig
	AssetConfigsMapByIndex           map[int]*types.AssetConfig
	AssetConfigsMapByProvider        map[string][]*types.AssetConfig
}

type Opt func(config *types.AppConfig)

func New(configUrl string, opt ...Opt) (ConfigDiscovery, error) {
	cfg := &configDiscovery{cfgUri: configUrl, Updates: make(chan *types.AppConfig)}

	for _, o := range opt {
		o(cfg.Config)
	}

	if err := cfg.FetchConfig(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *configDiscovery) ListenUpdates() error {
	if err := c.FetchConfig(); err != nil {
		return errors.Wrap(err, "update config")
	}

	if c.Config == nil {
		panic("Failed to fetch config from config discovery service")
	}

	go func() {
		for range time.Tick(time.Second * 5) {
			err := c.FetchConfig()

			if err != nil {
				log.Error().Err(err).Msg("update config err")
			}
		}
	}()

	return nil
}

func (c *configDiscovery) FetchConfig() error {
	cfg, err := request.Get[types.AppConfig](c.cfgUri)
	if err != nil {
		return errors.Wrap(err, "get app config")
	}

	if c.LastUpdatedAt == nil || cfg.ComposedAt != *c.LastUpdatedAt {
		log.Info().Msg("Config is updated, fetching updates")

		c.LastUpdatedAt = &cfg.ComposedAt

		assets, err := request.Get[[]*types.Asset](c.cfgUri + "/assets")
		if err != nil {
			return errors.Wrap(err, "fetch assets list")
		}

		schedule, err := request.Get[types.AssetsSchedule](c.cfgUri + "/assets-schedule")
		if err != nil {
			return errors.Wrap(err, "fetch assets schedule config")
		}

		conf, err := request.Get[[]*types.AssetConfig](c.cfgUri + "/assets-config")
		if err != nil {
			return errors.Wrap(err, "fetch assets config")
		}

		c.Config = &cfg
		c.Assets = assets
		c.AssetConfigs = conf
		c.Schedules = schedule.Schedules

		c.VaultsMapByAddress = make(map[string]*types.Vault)
		c.VaultsMapByCollateralAssetName = make(map[string]*types.Vault)
		c.VaultsMapByCollateralAssetId = make(map[string]*types.Vault)
		c.VaultsMapByLpJettonMasterAddress = make(map[string]*types.Vault)

		for _, v := range c.Config.Vaults {
			c.VaultsMapByAddress[v.VaultAddress] = &v
			c.VaultsMapByCollateralAssetName[v.Asset.Name] = &v
			c.VaultsMapByCollateralAssetId[v.Asset.AssetId] = &v
			c.VaultsMapByLpJettonMasterAddress[v.LpJettonMaster] = &v
		}

		c.MarketsMapByAddress = make(map[string]*types.Market)
		c.PrelaunchMarketsMapByAddress = make(map[string]*types.Market)
		c.MarketsMapByBaseAssetName = make(map[string][]types.Market)

		for _, m := range c.Config.OpenedMarkets {
			c.MarketsMapByAddress[m.Address] = &m
			c.MarketsMapByAddress[m.BaseAsset] = &m
			if m.Type == "prelaunch" {
				c.PrelaunchMarketsMapByAddress[m.Address] = &m
			}
			if c.MarketsMapByBaseAssetName[m.BaseAsset] == nil {
				c.MarketsMapByBaseAssetName[m.BaseAsset] = make([]types.Market, 0)
			}
			c.MarketsMapByBaseAssetName[m.BaseAsset] = append(c.MarketsMapByBaseAssetName[m.BaseAsset], m)
		}

		c.CollateralAssetsMapByName = make(map[string]*types.CollateralAsset)

		for _, a := range c.Config.CollateralAssets {
			c.CollateralAssetsMapByName[a.Name] = &a
		}

		c.AssetsMapByName = make(map[string]*types.Asset)
		c.AssetsMapByIndex = make(map[int]*types.Asset)

		for _, a := range c.Assets {
			c.AssetsMapByName[a.Name] = a
			c.AssetsMapByIndex[a.Index] = a
		}

		c.AssetConfigsMapByProvider = make(map[string][]*types.AssetConfig)
		c.AssetConfigsMapByName = make(map[string]*types.AssetConfig)
		for _, a := range c.AssetConfigs {
			c.AssetConfigsMapByName[a.Name] = a
			for _, o := range a.Oracles {
				if c.AssetConfigsMapByProvider[o.Provider] == nil {
					c.AssetConfigsMapByProvider[o.Provider] = make([]*types.AssetConfig, 0)
				}
				c.AssetConfigsMapByProvider[o.Provider] = append(c.AssetConfigsMapByProvider[o.Provider], a)
			}
		}

		go func() {
			c.Updates <- c.Config
		}()
	}

	return nil
}

func (c *configDiscovery) UpdatesChannel() <-chan *types.AppConfig {
	return c.Updates
}

func (c *configDiscovery) GetConfig() *types.AppConfig {
	return c.Config
}

func (c *configDiscovery) GetAssets() []*types.Asset {
	return c.Assets
}

func (c *configDiscovery) GetAssetConfigs() []*types.AssetConfig {
	return c.AssetConfigs
}

func (c *configDiscovery) GetSchedules() map[string]*types.AssetSchedule {
	return c.Schedules
}

func (c *configDiscovery) HasMarketByAddress(address string) bool {
	return c.MarketsMapByAddress[address] != nil
}

func (c *configDiscovery) GetMarketByAddress(address string) *types.Market {
	return c.MarketsMapByAddress[address]
}

func (c *configDiscovery) HasPrelaunchMarketByAddress(address string) bool {
	return c.PrelaunchMarketsMapByAddress[address] != nil
}

func (c *configDiscovery) GetPrelaunchMarketByAddress(address string) *types.Market {
	return c.PrelaunchMarketsMapByAddress[address]
}

func (c *configDiscovery) GetMarketsAddresses() []string {
	return maps.Keys(c.MarketsMapByAddress)
}

func (c *configDiscovery) GetMarketsByAssetName(name string) []types.Market {
	return c.MarketsMapByBaseAssetName[name]
}

func (c *configDiscovery) HasVaultByAddress(address string) bool {
	return c.VaultsMapByAddress[address] != nil
}

func (c *configDiscovery) GetVaultByAddress(address string) *types.Vault {
	return c.VaultsMapByAddress[address]
}

func (c *configDiscovery) HasVaultByLpJettonMasterAddress(address string) bool {
	return c.VaultsMapByLpJettonMasterAddress[address] != nil
}

func (c *configDiscovery) GetVaultByLpJettonMasterAddress(address string) *types.Vault {
	return c.VaultsMapByLpJettonMasterAddress[address]
}

func (c *configDiscovery) HasAssetByIndex(index int) bool {
	return c.AssetsMapByIndex[index] != nil
}

func (c *configDiscovery) GetAssetByIndex(index int) *types.Asset {
	return c.AssetsMapByIndex[index]
}

func (c *configDiscovery) HasAssetByName(name string) bool {
	return c.AssetsMapByName[name] != nil
}

func (c *configDiscovery) GetAssetByName(name string) *types.Asset {

	return c.AssetsMapByName[name]
}

func (c *configDiscovery) HasCollateralAssetByName(name string) bool {
	return c.CollateralAssetsMapByName[name] != nil
}

func (c *configDiscovery) GetCollateralAssetByName(name string) *types.CollateralAsset {
	return c.CollateralAssetsMapByName[name]
}

func (c *configDiscovery) HasVaultByCollateralAssetId(assetId string) bool {
	return c.VaultsMapByCollateralAssetId[assetId] != nil
}

func (c *configDiscovery) GetVaultByCollateralAssetId(assetId string) *types.Vault {
	return c.VaultsMapByCollateralAssetId[assetId]
}

func (c *configDiscovery) HasVaultByCollateralAssetName(name string) bool {
	return c.VaultsMapByCollateralAssetName[name] != nil
}

func (c *configDiscovery) GetVaultByCollateralAssetName(name string) *types.Vault {
	return c.VaultsMapByCollateralAssetName[name]
}

func (c *configDiscovery) HasAssetConfigByName(name string) bool {
	return c.AssetConfigsMapByName[name] != nil
}

func (c *configDiscovery) GetAssetConfigByName(name string) *types.AssetConfig {
	return c.AssetConfigsMapByName[name]
}

func (c *configDiscovery) HasAssetConfigByIndex(index int) bool {
	return c.AssetConfigsMapByIndex[index] != nil
}

func (c *configDiscovery) GetAssetConfigByIndex(index int) *types.AssetConfig {
	return c.AssetConfigsMapByIndex[index]
}

func (c *configDiscovery) GetAssetConfigsByProvider(name string) []*types.AssetConfig {
	return c.AssetConfigsMapByProvider[name]
}
