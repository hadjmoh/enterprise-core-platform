package enrichment

import (
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/pkg/logger"
)

// Pipeline orchestrates multiple enrichment services
type Pipeline struct {
	geoip  *GeoIPEnricher
	threat *ThreatIntelEnricher
	asset  *AssetEnricher
	user   *UserEnricher
	logger *logger.Logger
}

func NewPipeline(logger *logger.Logger) *Pipeline {
	return &Pipeline{
		threat: NewThreatIntelEnricher(logger),
		asset:  NewAssetEnricher(logger),
		user:   NewUserEnricher(logger),
		logger: logger,
	}
}

func (p *Pipeline) SetGeoIP(geoip *GeoIPEnricher) {
	p.geoip = geoip
}

func (p *Pipeline) Handle(event buffer.Event) error {
	data := event.Data

	// Apply enrichments in order
	if p.geoip != nil {
		data = p.geoip.Enrich(data)
	}
	
	if p.threat != nil {
		data = p.threat.Enrich(data)
	}
	
	if p.asset != nil {
		data = p.asset.Enrich(data)
	}
	
	if p.user != nil {
		data = p.user.Enrich(data)
	}

	// Update event with enriched data
	event.Data = data

	p.logger.Info("Event enriched", "source", event.Source)
	return nil
}

func (p *Pipeline) GetThreatEnricher() *ThreatIntelEnricher {
	return p.threat
}

func (p *Pipeline) GetAssetEnricher() *AssetEnricher {
	return p.asset
}

func (p *Pipeline) GetUserEnricher() *UserEnricher {
	return p.user
}
