package client

import (
	"math/big"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/storm-trade/config-discovery-client/request"
	"github.com/storm-trade/config-discovery-client/types"
	"golang.org/x/exp/maps"
)

type ConfigDiscovery interface {
	ListenUpdates() error
	GetConfig() *types.AppConfig
	GetAssets() []*types.Asset
	GetAssetConfigs() []*types.AssetConfig
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
	HasAssetConfigByName(name string) bool
	GetAssetConfigByName(name string) *types.AssetConfig
	HasAssetConfigByIndex(index int) bool
	GetAssetConfigByIndex(index int) *types.AssetConfig
	GetAssetConfigsByProvider(name string) []*types.AssetConfig
	IsLazer(address string) bool
	GetVPIHistory(name string) (map[int64]types.VPIParamsParsed, bool)
	GetVPIParamsAtTimestamp(name string, ts int64) (*types.VPIParamsParsed, bool)
	UpdatesChannel() <-chan *types.AppConfig
}

func strToBigInt(str string) (*big.Int, bool) {
	n := new(big.Int)
	return n.SetString(str, 10)
}

type configDiscovery struct {
	cfgUri        string
	Updates       chan *types.AppConfig
	LastUpdatedAt *string
	Config        *types.AppConfig
	Assets        []*types.Asset
	AssetConfigs  []*types.AssetConfig
	Schedules     map[string]*types.AssetSchedule

	VPIHistory map[string]map[int64]types.VPIParamsParsed
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
	LazerAssetsMap                   map[string]bool
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

		history, err := request.Get[map[string]map[string]types.VPIParams](c.cfgUri + "/vpi-history")
		if err != nil {
			return errors.Wrap(err, "fetch vpi history")
		}

		Config := &cfg
		Assets := assets
		AssetConfigs := conf
		Schedules := schedule.Schedules
		VPIHistory := make(map[string]map[int64]types.VPIParamsParsed)
		for name, h := range history {
			VPIHistory[name] = make(map[int64]types.VPIParamsParsed)
			for ts, params := range h {
				if params.MarketDepthLong == "" || params.MarketDepthShort == "" {
					continue
				}
				timestamp, err := strconv.ParseInt(ts, 10, 64)
				if err != nil {
					return errors.Wrap(err, "parse vpi timestamp")
				}
				marketDepthLong, ok := strToBigInt(params.MarketDepthLong)
				if !ok {
					return errors.New("parse vpi marketDepthLong")
				}
				marketDepthShort, ok := strToBigInt(params.MarketDepthShort)
				if !ok {
					return errors.New("parse vpi marketDepthShort")
				}
				spread, ok := strToBigInt(params.Spread)
				if !ok {
					return errors.New("parse vpi spread")
				}
				k, ok := strToBigInt(params.K)
				if !ok {
					return errors.New("parse vpi k")
				}
				VPIHistory[name][timestamp] = types.VPIParamsParsed{
					MarketDepthLong:  marketDepthLong,
					MarketDepthShort: marketDepthShort,
					Spread:           spread,
					K:                k,
				}
			}
		}

		VaultsMapByAddress := make(map[string]*types.Vault)
		VaultsMapByCollateralAssetName := make(map[string]*types.Vault)
		VaultsMapByCollateralAssetId := make(map[string]*types.Vault)
		VaultsMapByLpJettonMasterAddress := make(map[string]*types.Vault)

		for _, v := range Config.Vaults {
			VaultsMapByAddress[v.VaultAddress] = &v
			VaultsMapByCollateralAssetName[v.Asset.Name] = &v
			VaultsMapByCollateralAssetId[v.Asset.AssetId] = &v
			VaultsMapByLpJettonMasterAddress[v.LpJettonMaster] = &v
		}

		MarketsMapByAddress := make(map[string]*types.Market)
		PrelaunchMarketsMapByAddress := make(map[string]*types.Market)
		MarketsMapByBaseAssetName := make(map[string][]types.Market)

		for _, m := range Config.OpenedMarkets {
			MarketsMapByAddress[m.Address] = &m
			MarketsMapByAddress[m.BaseAsset] = &m
			if m.Type == "prelaunch" {
				PrelaunchMarketsMapByAddress[m.Address] = &m
			}
			if MarketsMapByBaseAssetName[m.BaseAsset] == nil {
				MarketsMapByBaseAssetName[m.BaseAsset] = make([]types.Market, 0)
			}
			MarketsMapByBaseAssetName[m.BaseAsset] = append(MarketsMapByBaseAssetName[m.BaseAsset], m)
		}

		CollateralAssetsMapByName := make(map[string]*types.CollateralAsset)

		for _, a := range Config.CollateralAssets {
			CollateralAssetsMapByName[a.Name] = &a
		}

		AssetsMapByName := make(map[string]*types.Asset)
		AssetsMapByIndex := make(map[int]*types.Asset)

		for _, a := range Assets {
			AssetsMapByName[a.Name] = a
			AssetsMapByIndex[a.Index] = a
		}

		AssetConfigsMapByProvider := make(map[string][]*types.AssetConfig)
		AssetConfigsMapByName := make(map[string]*types.AssetConfig)
		AssetConfigsMapByIndex := make(map[int]*types.AssetConfig)
		LazerAssetsMap := make(map[string]bool)
		for _, a := range AssetConfigs {
			AssetConfigsMapByName[a.Name] = a
			AssetConfigsMapByIndex[a.Index] = a
			for _, o := range a.Oracles {
				if o.Provider == "pyth-lazer" || o.Provider == "stork-fast" || o.Provider == "fake" {
					LazerAssetsMap[a.Name] = true
				}
				if AssetConfigsMapByProvider[o.Provider] == nil {
					AssetConfigsMapByProvider[o.Provider] = make([]*types.AssetConfig, 0)
				}
				AssetConfigsMapByProvider[o.Provider] = append(AssetConfigsMapByProvider[o.Provider], a)
			}
		}

		{
			c.Config = Config
			c.Assets = Assets
			c.AssetConfigs = AssetConfigs
			c.Schedules = Schedules
			c.VPIHistory = VPIHistory
			c.VaultsMapByAddress = VaultsMapByAddress
			c.VaultsMapByCollateralAssetName = VaultsMapByCollateralAssetName
			c.VaultsMapByCollateralAssetId = VaultsMapByCollateralAssetId
			c.VaultsMapByLpJettonMasterAddress = VaultsMapByLpJettonMasterAddress
			c.MarketsMapByAddress = MarketsMapByAddress
			c.PrelaunchMarketsMapByAddress = PrelaunchMarketsMapByAddress
			c.MarketsMapByBaseAssetName = MarketsMapByBaseAssetName
			c.AssetsMapByName = AssetsMapByName
			c.AssetsMapByIndex = AssetsMapByIndex
			c.CollateralAssetsMapByName = CollateralAssetsMapByName
			c.AssetConfigsMapByName = AssetConfigsMapByName
			c.AssetConfigsMapByIndex = AssetConfigsMapByIndex
			c.AssetConfigsMapByProvider = AssetConfigsMapByProvider
			c.LazerAssetsMap = LazerAssetsMap
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

func (c *configDiscovery) GetVPIHistory(name string) (map[int64]types.VPIParamsParsed, bool) {
	i, ok := c.VPIHistory[name]
	return i, ok
}

func (c *configDiscovery) GetVPIParamsAtTimestamp(name string, ts int64) (*types.VPIParamsParsed, bool) {
	i, ok := c.VPIHistory[name]
	if !ok {
		return nil, false
	}
	// It's sorted in reverse timestamp
	for timestamp, params := range i {
		if timestamp <= ts {
			return &params, true
		}
	}
	return nil, false
}

func (c *configDiscovery) IsLazer(name string) bool {
	_, ok := c.LazerAssetsMap[name]
	return ok
}
