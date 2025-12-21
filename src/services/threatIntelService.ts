import { v4 as uuidv4 } from 'uuid';

export interface TiIndicator {
  id: string;
  type: 'indicator';
  pattern: string; // "[ipv4-addr:value = '198.51.100.1']"
  valid_from: string;
  valid_until?: string;
  confidence: number; // 0-100
  severity: 'low' | 'medium' | 'high' | 'critical';
  labels: string[]; // ["malicious-activity", "ransomware"]
  source: string; // "AlienVault", "CI-Army"
  created: string;
}

export interface TiFeed {
  id: string;
  name: string;
  description: string;
  status: 'active' | 'disconnected' | 'syncing';
  lastSync: string | null;
  indicatorCount: number;
}

class ThreatIntelService {
  private indicators: Map<string, TiIndicator> = new Map();
  private feeds: TiFeed[] = [
    {
      id: 'feed-alienvault',
      name: 'AlienVault OTX',
      description: 'Open Threat Exchange community feed (Mock)',
      status: 'disconnected',
      lastSync: null,
      indicatorCount: 0
    },
    {
      id: 'feed-ciarmy',
      name: 'CI Army Bad IPs',
      description: 'Active attackers IP list (Mock)',
      status: 'disconnected',
      lastSync: null,
      indicatorCount: 0
    },
    {
      id: 'feed-phish',
      name: 'OpenPhish',
      description: 'Zero-day phishing sites (Mock)',
      status: 'disconnected',
      lastSync: null,
      indicatorCount: 0
    }
  ];

  constructor() {
    this.loadFromStorage();
  }

  // --- Persistence ---
  private saveToStorage() {
    localStorage.setItem('ti_feeds', JSON.stringify(this.feeds));
    // In a real app, indicators would be in a DB. For mock, we'll keep top 1000 in memory/localstorage or just memory to avoid quota issues.
    // Let's rely on memory for frequent changes, but save feeds state.
  }

  private loadFromStorage() {
    const saved = localStorage.getItem('ti_feeds');
    if (saved) {
      this.feeds = JSON.parse(saved);
    }
  }

  // --- Public API ---

  getFeeds(): TiFeed[] {
    return [...this.feeds];
  }

  getStats() {
    const total = this.indicators.size;
    const critical = Array.from(this.indicators.values()).filter(i => i.severity === 'critical').length;
    const high = Array.from(this.indicators.values()).filter(i => i.severity === 'high').length;
    
    // Group by type (heuristic)
    const types = {
      ip: 0,
      domain: 0,
      hash: 0
    };
    
    this.indicators.forEach(i => {
      if (i.pattern.includes('ipv4-addr')) types.ip++;
      else if (i.pattern.includes('domain-name')) types.domain++;
      else if (i.pattern.includes('file:hashes')) types.hash++;
    });

    return { total, critical, high, types };
  }

  getRecentIndicators(limit = 50): TiIndicator[] {
    return Array.from(this.indicators.values())
      .sort((a, b) => new Date(b.created).getTime() - new Date(a.created).getTime())
      .slice(0, limit);
  }

  // --- Ingestion Simulation ---

  async ingestFeed(feedId: string): Promise<number> {
    const feed = this.feeds.find(f => f.id === feedId);
    if (!feed) throw new Error('Feed not found');

    feed.status = 'syncing';
    this.saveToStorage(); // Trigger UI update if polling, usually we'd use a listener

    // Simulate network delay
    await new Promise(r => setTimeout(r, 1500 + Math.random() * 1000));

    // Generate Mock Artifacts
    const newCount = 50 + Math.floor(Math.random() * 150);
    const newIndicators = this.generateMockIndicators(newCount, feed.name);

    newIndicators.forEach(i => this.indicators.set(i.id, i));

    feed.status = 'active';
    feed.lastSync = new Date().toISOString();
    feed.indicatorCount += newCount;
    this.saveToStorage();

    return newCount;
  }

  // --- Enrichment & Search ---

  // Fast lookup for O(1) correlation
  enrichArtifact(value: string): TiIndicator | null {
    // Naive linear scan for this mock since patterns are strings like "[ipv4...]"
    // In prod, this would be an inverted index or bloom filter.
    for (const indicator of this.indicators.values()) {
      if (indicator.pattern.includes(value)) {
        return indicator;
      }
    }
    return null;
  }

  manualSearch(query: string): TiIndicator[] {
    if (!query) return [];
    const lowerQ = query.toLowerCase();
    return Array.from(this.indicators.values()).filter(i => 
      i.pattern.toLowerCase().includes(lowerQ) || 
      i.labels.some(l => l.toLowerCase().includes(lowerQ))
    );
  }

  // --- Helpers ---

  private generateMockIndicators(count: number, source: string): TiIndicator[] {
    const results: TiIndicator[] = [];
    const severities: TiIndicator['severity'][] = ['low', 'medium', 'high', 'critical'];
    const attackTypes = ['ransomware', 'c2-server', 'phishing', 'botnet', 'apt-activity', 'scanner'];

    for (let i = 0; i < count; i++) {
        const typeRoll = Math.random();
        let pattern = '';
        if (typeRoll < 0.6) {
             // IP
             pattern = `[ipv4-addr:value = '${this.randomIP()}']`;
        } else if (typeRoll < 0.9) {
             // Domain
             pattern = `[domain-name:value = '${this.randomDomain()}']`;
        } else {
             // Hash
             pattern = `[file:hashes.'SHA-256' = '${this.randomHash()}']`;
        }

        results.push({
            id: `indicator--${uuidv4()}`,
            type: 'indicator',
            pattern,
            valid_from: new Date().toISOString(),
            confidence: Math.floor(Math.random() * 50) + 50, // 50-100
            severity: severities[Math.floor(Math.random() * severities.length)],
            labels: [attackTypes[Math.floor(Math.random() * attackTypes.length)]],
            source: source,
            created: new Date().toISOString()
        });
    }
    return results;
  }

  private randomIP() {
      return `${Math.floor(Math.random()*255)}.${Math.floor(Math.random()*255)}.${Math.floor(Math.random()*255)}.${Math.floor(Math.random()*255)}`;
  }

  private randomDomain() {
      const domains = ['evil.com', 'phish.net', 'malware.io', 'c2.xyz', 'apt29.org', 'steal-creds.co', 'crypto-miner.pool'];
      return `${this.randomString(5)}.${domains[Math.floor(Math.random() * domains.length)]}`;
  }

  private randomHash() {
      return Array.from({length: 64}, () => Math.floor(Math.random()*16).toString(16)).join('');
  }

  private randomString(len: number) {
      return Math.random().toString(36).substring(2, 2+len);
  }
}

export const threatIntelService = new ThreatIntelService();
