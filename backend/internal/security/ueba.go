package security

import (
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/pkg/logger"
	"sync"
	"time"
)

// Anomaly represents a detected behavioral deviation
type Anomaly struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
	EntityID    string    `json:"entity_id"`
}

// BehaviorProfile tracks established patterns for an entity
type BehaviorProfile struct {
	EntityID      string               `json:"entity_id"`
	SeenIPs       map[string]time.Time `json:"seen_ips"`
	SeenHours     [24]int             `json:"seen_hours"`
	CommonProc    map[string]int      `json:"common_proc"`
	LastLocation  string               `json:"last_location"` // Mocked as IP for now
	LastTimestamp time.Time            `json:"last_timestamp"`
	mu            sync.RWMutex
}

// UEBAEngine handles behavioral baselining and anomaly detection
type UEBAEngine struct {
	profiles map[string]*BehaviorProfile
	logger   *logger.Logger
	mu       sync.RWMutex
}

func NewUEBAEngine(logg *logger.Logger) *UEBAEngine {
	return &UEBAEngine{
		profiles: make(map[string]*BehaviorProfile),
		logger:   logg,
	}
}

func (e *UEBAEngine) GetProfile(entityID string) *BehaviorProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	p, ok := e.profiles[entityID]
	if !ok {
		p = &BehaviorProfile{
			EntityID:   entityID,
			SeenIPs:    make(map[string]time.Time),
			CommonProc: make(map[string]int),
		}
		e.profiles[entityID] = p
	}
	return p
}

func (e *UEBAEngine) Process(event buffer.Event) []Anomaly {
	entityID := ""
	if user, ok := event.Data["user"].(string); ok {
		entityID = user
	} else if srcIP, ok := event.Data["src_ip"].(string); ok {
		entityID = srcIP
	}

	if entityID == "" {
		return nil
	}

	profile := e.GetProfile(entityID)
	anomalies := e.detectAnomalies(profile, event)
	e.updateProfile(profile, event)

	return anomalies
}

func (e *UEBAEngine) detectAnomalies(p *BehaviorProfile, event buffer.Event) []Anomaly {
	p.mu.RLock()
	defer p.mu.RUnlock()

	anomalies := make([]Anomaly, 0)

	// 1. After-Hours Activity
	t, err := time.Parse(time.RFC3339, event.Timestamp)
	if err == nil {
		hour := t.Hour()
		// If this hour is in top 20% of quietest hours for user, flag it
		totalActivity := 0
		for _, count := range p.SeenHours {
			totalActivity += count
		}
		if totalActivity > 50 { // Only after baseline is established
			if p.SeenHours[hour] < (totalActivity / 48) { // Less than 2% of activity
				anomalies = append(anomalies, Anomaly{
					Type:        "AfterHoursActivity",
					Description: "Activity detected during atypical hours for this user.",
					Severity:    "medium",
					Timestamp:   time.Now(),
					EntityID:    p.EntityID,
				})
			}
		}
	}

	// 2. Impossible Travel
	if srcIP, ok := event.Data["src_ip"].(string); ok {
		if p.LastLocation != "" && p.LastLocation != srcIP {
			// Mocked Geography: Logic would calculate distance between IPs
			// For now: If IP changes within 5 minutes, flag as impossible travel
			if time.Since(p.LastTimestamp) < 5*time.Minute {
				anomalies = append(anomalies, Anomaly{
					Type:        "ImpossibleTravel",
					Description: "Rapid location change detected between disparate IP addresses.",
					Severity:    "high",
					Timestamp:   time.Now(),
					EntityID:    p.EntityID,
				})
			}
		}
	}

	// 3. Rare Process
	if proc, ok := event.Data["process_name"].(string); ok {
		if count, ok := p.CommonProc[proc]; ok {
			totalProc := 0
			for _, c := range p.CommonProc {
				totalProc += c
			}
			if totalProc > 100 && count < (totalProc/100) { // < 1% of baseline
				anomalies = append(anomalies, Anomaly{
					Type:        "RareProcess",
					Description: "Execution of a process rarely seen on this entity.",
					Severity:    "medium",
					Timestamp:   time.Now(),
					EntityID:    p.EntityID,
				})
			}
		} else if len(p.CommonProc) > 10 { // Only flag if we have a baseline
			anomalies = append(anomalies, Anomaly{
				Type:        "RareProcess",
				Description: "Execution of a process never before seen on this entity.",
				Severity:    "high",
				Timestamp:   time.Now(),
				EntityID:    p.EntityID,
			})
		}
	}

	return anomalies
}

func (e *UEBAEngine) updateProfile(p *BehaviorProfile, event buffer.Event) {
	p.mu.Lock()
	defer p.mu.Unlock()

	t, err := time.Parse(time.RFC3339, event.Timestamp)
	if err == nil {
		p.SeenHours[t.Hour()]++
		p.LastTimestamp = t
	}

	if srcIP, ok := event.Data["src_ip"].(string); ok {
		p.SeenIPs[srcIP] = time.Now()
		p.LastLocation = srcIP
	}

	if proc, ok := event.Data["process_name"].(string); ok {
		p.CommonProc[proc]++
	}
}

func (e *UEBAEngine) GetGlobalAnomalies() []Anomaly {
	// Implementation for global aggregate if needed
	return nil
}

func (e *UEBAEngine) GetProfileSummary(entityID string) map[string]interface{} {
	p := e.GetProfile(entityID)
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := 0
	for _, c := range p.SeenHours {
		total += c
	}

	return map[string]interface{}{
		"entity_id":      p.EntityID,
		"total_activity": total,
		"hour_baseline":  p.SeenHours,
		"unique_ips":     len(p.SeenIPs),
		"unique_procs":   len(p.CommonProc),
	}
}
