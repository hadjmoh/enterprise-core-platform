package api



// Event represents the standardized Splunk-compatible HEC event structure
type Event struct {
	Time       int64                  `json:"time"`                 // Epoch timestamp
	Host       string                 `json:"host,omitempty"`       // Originating host
	Source     string                 `json:"source,omitempty"`     // File path or stream name
	SourceType string                 `json:"sourcetype,omitempty"` // Format (e.g., access_combined, json)
	Index      string                 `json:"index,omitempty"`      // Destination index
	Event      interface{}            `json:"event"`                // Result payload (string or JSON object)
	Fields     map[string]interface{} `json:"fields,omitempty"`     // Custom fields
}

// HECResponse is the standard response for the collector
type HECResponse struct {
	Text string `json:"text"`
	Code int    `json:"code"`
}
