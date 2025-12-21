package enrichment

import (
	"context"
	"enterprise-core/backend/internal/auth"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/pkg/logger"
)

// Pipeline orchestrates multiple enrichment services
type Pipeline struct {
	geoip    *GeoIPEnricher
	threat   *ThreatIntelEnricher
	asset    *AssetEnricher
	user     *UserEnricher
	identity auth.IdentityResolver
	logger   *logger.Logger
}

func NewPipeline(logger *logger.Logger, identity auth.IdentityResolver) *Pipeline {
	return &Pipeline{
		threat:   NewThreatIntelEnricher(logger),
		asset:    NewAssetEnricher(logger),
		user:     NewUserEnricher(logger),
		identity: identity,
		logger:   logger,
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

	// Dynamic Identity Resolution & Mapping (Unified Identity Phase)
	if p.identity != nil {
		identifier := ""
		if user, ok := data["user"].(string); ok {
			identifier = user
		} else if ip, ok := data["src_ip"].(string); ok {
			identifier = ip
		}

		if identifier != "" {
			if id, err := p.identity.Resolve(context.Background(), identifier); err == nil {
				data["owner_user"] = id.PrimaryUser
				data["owner_department"] = id.Department
				data["owner_uid"] = id.UID
			}
		}
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
