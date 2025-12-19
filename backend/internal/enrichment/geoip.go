package enrichment

import (
	"enterprise-core/backend/pkg/logger"
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

type GeoIPEnricher struct {
	db     *geoip2.Reader
	logger *logger.Logger
	mu     sync.RWMutex
}

func NewGeoIPEnricher(dbPath string, logger *logger.Logger) (*GeoIPEnricher, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}

	return &GeoIPEnricher{
		db:     db,
		logger: logger,
	}, nil
}

func (g *GeoIPEnricher) Enrich(data map[string]interface{}) map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Extract IP addresses from common fields
	ipFields := []string{"src_ip", "dst_ip", "client_ip", "remote_addr", "ip"}
	
	for _, field := range ipFields {
		if ipStr, ok := data[field].(string); ok {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				continue
			}

			record, err := g.db.City(ip)
			if err != nil {
				continue
			}

			// Add geo data
			prefix := field + "_geo"
			data[prefix+"_country"] = record.Country.IsoCode
			data[prefix+"_city"] = record.City.Names["en"]
			data[prefix+"_latitude"] = record.Location.Latitude
			data[prefix+"_longitude"] = record.Location.Longitude
			
			if len(record.Subdivisions) > 0 {
				data[prefix+"_region"] = record.Subdivisions[0].Names["en"]
			}
		}
	}

	return data
}

func (g *GeoIPEnricher) Close() error {
	return g.db.Close()
}
