package query

import (
	"context"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/compliance"
	"enterprise-core/backend/internal/storage"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SearchProcessor is the primary source command for every pipeline
type SearchProcessor struct {
	engine  *storage.FileStorageEngine
	query   storage.Query
	reader  storage.StorageReader
	metrics CommandMetrics
}

func NewSearchProcessor(engine *storage.FileStorageEngine, node CommandNode) (*SearchProcessor, error) {
	q := storage.Query{
		StartTime: time.Now().Add(-24 * time.Hour), // Default to last 24h
		EndTime:   time.Now(),
		Limit:     1000, // Default limit
	}

	for _, arg := range node.Args {
		if arg.Key == "" {
			q.Keywords = append(q.Keywords, arg.Value)
		} else {
			if q.FieldFilters == nil {
				q.FieldFilters = make(map[string]string)
			}
			q.FieldFilters[arg.Key] = arg.Value
			
			// Special handling for time-range helpers if needed (not implemented yet)
		}
	}

	return &SearchProcessor{
		engine: engine,
		query:  q,
	}, nil
}

func (p *SearchProcessor) Init(ctx context.Context) error {
	reader, err := p.engine.NewReader(ctx, p.query)
	if err != nil {
		return fmt.Errorf("search failed to create reader: %w", err)
	}
	p.reader = reader
	return nil
}

func (p *SearchProcessor) Close() error {
	if p.reader != nil {
		return p.reader.Close()
	}
	return nil
}

func (p *SearchProcessor) Metrics() CommandMetrics {
	return p.metrics
}

func (p *SearchProcessor) Process(ctx context.Context, in <-chan buffer.Event, out chan<- buffer.Event) error {
	// Search ignores 'in' channel as it is a source
	
	if p.reader == nil {
		return fmt.Errorf("search processor not initialized")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ev, err := p.reader.Next(ctx)
			if err != nil {
				if err.Error() == "EOF" {
					return nil
				}
				return err
			}
			
			// Filtering logic for the SPL Search command
			if p.matches(ev) {
				out <- ev
				p.metrics.EventsOut++
			}
		}
	}
}

func (p *SearchProcessor) matches(ev buffer.Event) bool {
	// Keyword matching
	if len(p.query.Keywords) > 0 {
		msg, ok := ev.Data["message"].(string)
		if !ok {
			return false
		}
		for _, kw := range p.query.Keywords {
			// Basic case-insensitive match for keywords in the message
			// This matches our implementation in FileStorageEngine.Search
			found := false
			// Simple check for now
			if strings.Contains(strings.ToLower(msg), strings.ToLower(kw)) {
				found = true
			}
			if !found {
				return false
			}
		}
	}

	// Field matching
	for k, v := range p.query.FieldFilters {
		val, ok := ev.Data[k]
		if !ok || fmt.Sprintf("%v", val) != v {
			return false
		}
	}

	return true
}

// Predicate represents a single condition in a 'where' clause
type Predicate struct {
	LeftField  string
	Operator   string
	RightValue string
	RightField string // If empty, we compare against RightValue
	IsNegated  bool
}

// WhereProcessor performs conditional filtering with support for field comparisons and negation
type WhereProcessor struct {
	predicates []Predicate
	isOr       bool
	metrics    CommandMetrics
}

func (p *WhereProcessor) Init(ctx context.Context) error { return nil }
func (p *WhereProcessor) Close() error                 { return nil }
func (p *WhereProcessor) Metrics() CommandMetrics     { return p.metrics }

func (p *WhereProcessor) Process(ctx context.Context, in <-chan buffer.Event, out chan<- buffer.Event) error {
	for ev := range in {
		p.metrics.EventsIn++
		if p.matches(ev) {
			out <- ev
			p.metrics.EventsOut++
		} else {
			p.metrics.EventsDropped++
		}
	}
	return nil
}

func (p *WhereProcessor) matches(ev buffer.Event) bool {
	if len(p.predicates) == 0 {
		return true
	}

	for _, pred := range p.predicates {
		match := p.evalPredicate(pred, ev)
		if pred.IsNegated {
			match = !match
		}

		if p.isOr {
			if match {
				return true
			}
		} else {
			if !match {
				return false
			}
		}
	}
	return !p.isOr
}

func (p *WhereProcessor) evalPredicate(pred Predicate, ev buffer.Event) bool {
	leftVal := p.getFieldValue(pred.LeftField, ev)
	if leftVal == nil {
		return pred.Operator == "!="
	}

	var rightVal interface{}
	if pred.RightField != "" {
		rightVal = p.getFieldValue(pred.RightField, ev)
	} else {
		rightVal = pred.RightValue
	}

	if rightVal == nil {
		return pred.Operator == "!="
	}

	s1 := fmt.Sprintf("%v", leftVal)
	s2 := fmt.Sprintf("%v", rightVal)

	switch pred.Operator {
	case "=", "==":
		return s1 == s2
	case "!=":
		return s1 != s2
	case ">":
		return s1 > s2
	case "<":
		return s1 < s2
	case ">=":
		return s1 >= s2
	case "<=":
		return s1 <= s2
	default:
		return false
	}
}

func (p *WhereProcessor) getFieldValue(field string, ev buffer.Event) interface{} {
	switch field {
	case "source":
		return ev.Source
	case "timestamp":
		return ev.Timestamp
	default:
		return ev.Data[field]
	}
}

// FieldsProcessor projects or removes fields, now with regex support
type FieldsProcessor struct {
	fieldSpecs []string // original order
	fields     map[string]bool
	regexes    []*regexp.Regexp
	isKeep     bool
	metrics    CommandMetrics
}

func (p *FieldsProcessor) Init(ctx context.Context) error { return nil }
func (p *FieldsProcessor) Close() error                 { return nil }
func (p *FieldsProcessor) Metrics() CommandMetrics     { return p.metrics }

func (p *FieldsProcessor) Process(ctx context.Context, in <-chan buffer.Event, out chan<- buffer.Event) error {
	for ev := range in {
		p.metrics.EventsIn++
		newData := make(map[string]interface{})

		if p.isKeep {
			// Handle regex and explicit fields
			for k, v := range ev.Data {
				if p.shouldKeep(k) {
					newData[k] = v
				}
			}
			// Special top-level handling
			if !p.shouldKeep("source") {
				ev.Source = ""
			}
			if !p.shouldKeep("timestamp") {
				ev.Timestamp = ""
			}
		} else {
			// Remove specific fields
			for k, v := range ev.Data {
				if !p.shouldRemove(k) {
					newData[k] = v
				}
			}
			if p.shouldRemove("source") {
				ev.Source = ""
			}
			if p.shouldRemove("timestamp") {
				ev.Timestamp = ""
			}
		}

		ev.Data = newData
		out <- ev
		p.metrics.EventsOut++
	}
	return nil
}

func (p *FieldsProcessor) shouldKeep(field string) bool {
	if p.fields[field] {
		return true
	}
	for _, re := range p.regexes {
		if re.MatchString(field) {
			return true
		}
	}
	return false
}

func (p *FieldsProcessor) shouldRemove(field string) bool {
	return p.shouldKeep(field) // Logic is inverted in Process
}

// EvalProcessor creates or transforms fields with arithmetic support
type EvalProcessor struct {
	targetField string
	op          string // +, -, *, / or empty for copy
	leftField   string
	rightField  string
	rightConst  float64
	hasConst    bool
	metrics     CommandMetrics
}

func (p *EvalProcessor) Init(ctx context.Context) error { return nil }
func (p *EvalProcessor) Close() error                 { return nil }
func (p *EvalProcessor) Metrics() CommandMetrics     { return p.metrics }

func (p *EvalProcessor) Process(ctx context.Context, in <-chan buffer.Event, out chan<- buffer.Event) error {
	for ev := range in {
		p.metrics.EventsIn++

		var result interface{}
		
		if p.op == "" {
			// Simple copy
			result = p.getVal(p.leftField, ev)
		} else {
			// Arithmetic
			v1 := p.getFloat(p.leftField, ev)
			var v2 float64
			if p.hasConst {
				v2 = p.rightConst
			} else {
				v2 = p.getFloat(p.rightField, ev)
			}

			switch p.op {
			case "+":
				result = v1 + v2
			case "-":
				result = v1 - v2
			case "*":
				result = v1 * v2
			case "/":
				if v2 != 0 {
					result = v1 / v2
				} else {
					result = nil // Safety: avoid div by zero
				}
			}
		}

		if result != nil {
			p.setVal(p.targetField, result, &ev)
		}

		out <- ev
		p.metrics.EventsOut++
	}
	return nil
}

func (p *EvalProcessor) getVal(field string, ev buffer.Event) interface{} {
	switch field {
	case "source":
		return ev.Source
	case "timestamp":
		return ev.Timestamp
	default:
		return ev.Data[field]
	}
}

func (p *EvalProcessor) getFloat(field string, ev buffer.Event) float64 {
	val := p.getVal(field, ev)
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0
	}
}

func (p *EvalProcessor) setVal(field string, val interface{}, ev *buffer.Event) {
	switch field {
	case "source":
		ev.Source = fmt.Sprintf("%v", val)
	case "timestamp":
		ev.Timestamp = fmt.Sprintf("%v", val)
	default:
		ev.Data[field] = val
	}
}

// LimitProcessor truncates the result stream
type LimitProcessor struct {
	limit   int
	count   int
	metrics CommandMetrics
}

func (p *LimitProcessor) Init(ctx context.Context) error {
	return nil
}

func (p *LimitProcessor) Close() error {
	return nil
}

func (p *LimitProcessor) Metrics() CommandMetrics {
	return p.metrics
}

func (p *LimitProcessor) Process(ctx context.Context, in <-chan buffer.Event, out chan<- buffer.Event) error {
	for ev := range in {
		p.metrics.EventsIn++
		if p.count < p.limit {
			out <- ev
			p.count++
			p.metrics.EventsOut++
		} else {
			p.metrics.EventsDropped++
			// Signal upstream could be done via context or simply stopping the drain
			// For linear pipes, stopping the drain will put backpressure on the previous stage
			break
		}
	}
	return nil
}

const (
	MaxGroupsDefault = 50000
	MaxSeriesDefault = 500
	MaxSpanDays      = 366
	KeySeparator     = "\x1f" // Unit Separator (non-printable)
)

// Aggregator interface for statistical functions
type Aggregator interface {
	Update(val interface{})
	Result() interface{}
}

type CountAggregator struct {
	count int64
}

func (a *CountAggregator) Update(val interface{}) { a.count++ }
func (a *CountAggregator) Result() interface{}    { return a.count }

type SumAggregator struct {
	sum float64
}

func (a *SumAggregator) Update(val interface{}) {
	if f, ok := toFloat(val); ok {
		a.sum += f
	}
}
func (a *SumAggregator) Result() interface{} { return a.sum }

type AvgAggregator struct {
	sum   float64
	count int64
}

func (a *AvgAggregator) Update(val interface{}) {
	if f, ok := toFloat(val); ok {
		a.sum += f
		a.count++
	}
}
func (a *AvgAggregator) Result() interface{} {
	if a.count == 0 {
		return 0.0
	}
	return a.sum / float64(a.count)
}

type MinAggregator struct {
	min   float64
	first bool
}

func (a *MinAggregator) Update(val interface{}) {
	if f, ok := toFloat(val); ok {
		if !a.first || f < a.min {
			a.min = f
			a.first = true
		}
	}
}
func (a *MinAggregator) Result() interface{} { return a.min }

type MaxAggregator struct {
	max   float64
	first bool
}

func (a *MaxAggregator) Update(val interface{}) {
	if f, ok := toFloat(val); ok {
		if !a.first || f > a.max {
			a.max = f
			a.first = true
		}
	}
}
func (a *MaxAggregator) Result() interface{} { return a.max }

func toFloat(val interface{}) (float64, bool) {
	if val == nil {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

// AggregationSpec defines what to aggregate and how
type AggregationSpec struct {
	Func  string
	Field string
	Alias string
}

// StatsProcessor performs global and grouped aggregations
type StatsProcessor struct {
	specs     []AggregationSpec
	byFields  []string
	groups    map[string]map[string]Aggregator // groupKey -> alias -> aggregator
	maxGroups int
	metrics   CommandMetrics
}

func NewStatsProcessor(specs []AggregationSpec, byFields []string, maxGroups int) *StatsProcessor {
	if maxGroups <= 0 {
		maxGroups = MaxGroupsDefault
	}
	return &StatsProcessor{
		specs:     specs,
		byFields:  byFields,
		groups:    make(map[string]map[string]Aggregator),
		maxGroups: maxGroups,
	}
}

func (p *StatsProcessor) Init(ctx context.Context) error { return nil }
func (p *StatsProcessor) Close() error                 { return nil }
func (p *StatsProcessor) Metrics() CommandMetrics     { return p.metrics }

func (p *StatsProcessor) Process(ctx context.Context, in <-chan buffer.Event, out chan<- buffer.Event) error {
	for ev := range in {
		p.metrics.EventsIn++
		
		// Generate group key
		var groupKey string
		if len(p.byFields) > 0 {
			var keys []string
			for _, f := range p.byFields {
				val := p.getFieldValue(f, ev)
				if val == nil {
					keys = append(keys, "(null)")
				} else {
					keys = append(keys, fmt.Sprintf("%v", val))
				}
			}
			groupKey = strings.Join(keys, KeySeparator)
		} else {
			groupKey = "global"
		}

		// Get or create group
		group, ok := p.groups[groupKey]
		if !ok {
			if len(p.groups) >= p.maxGroups {
				p.metrics.EventsDropped++
				continue // Skip record if group limit reached
			}
			group = make(map[string]Aggregator)
			for _, spec := range p.specs {
				group[spec.Alias] = p.createAggregator(spec.Func)
			}
			p.groups[groupKey] = group
		}

		// Update aggregators
		for _, spec := range p.specs {
			val := p.getFieldValue(spec.Field, ev)
			group[spec.Alias].Update(val)
		}
	}

	// Flush results
	for key, group := range p.groups {
		resEv := buffer.Event{
			Timestamp: time.Now().Format(time.RFC3339),
			Source:    "stats",
			Data:      make(map[string]interface{}),
		}

		// Add grouped fields back if applicable
		if len(p.byFields) > 0 {
			keyParts := strings.Split(key, KeySeparator)
			for i, f := range p.byFields {
				resEv.Data[f] = keyParts[i]
			}
		}

		// Add aggregated values
		for alias, agg := range group {
			resEv.Data[alias] = agg.Result()
		}

		out <- resEv
		p.metrics.EventsOut++
	}

	return nil
}

func (p *StatsProcessor) getFieldValue(field string, ev buffer.Event) interface{} {
	switch field {
	case "source":
		return ev.Source
	case "_time", "timestamp":
		return ev.Timestamp
	default:
		return ev.Data[field]
	}
}

func (p *StatsProcessor) createAggregator(fn string) Aggregator {
	switch strings.ToLower(fn) {
	case "count":
		return &CountAggregator{}
	case "sum":
		return &SumAggregator{}
	case "avg":
		return &AvgAggregator{}
	case "min":
		return &MinAggregator{}
	case "max":
		return &MaxAggregator{}
	default:
		return &CountAggregator{}
	}
}

// BucketProcessor aligns event timestamps to a specific interval
type BucketProcessor struct {
	span    time.Duration
	field   string // Usually _time or timestamp
	metrics CommandMetrics
}

func (p *BucketProcessor) Init(ctx context.Context) error { return nil }
func (p *BucketProcessor) Close() error                 { return nil }
func (p *BucketProcessor) Metrics() CommandMetrics     { return p.metrics }

func (p *BucketProcessor) Process(ctx context.Context, in <-chan buffer.Event, out chan<- buffer.Event) error {
	spanSecs := int64(p.span.Seconds())
	if spanSecs <= 0 {
		return fmt.Errorf("invalid bucket span: %v", p.span)
	}
	if spanSecs > MaxSpanDays*24*3600 {
		return fmt.Errorf("bucket span exceeds maximum limit of %d days", MaxSpanDays)
	}

	for ev := range in {
		p.metrics.EventsIn++
		
		t, err := time.Parse(time.RFC3339, ev.Timestamp)
		if err == nil {
			// Align to span in UTC
			t = t.UTC()
			bucketed := (t.Unix() / spanSecs) * spanSecs
			ev.Timestamp = time.Unix(bucketed, 0).UTC().Format(time.RFC3339)
			if ev.Data == nil {
				ev.Data = make(map[string]interface{})
			}
			ev.Data["_time"] = ev.Timestamp
			out <- ev
			p.metrics.EventsOut++
		} else {
			p.metrics.EventsDropped++
		}
	}
	return nil
}

// TimechartProcessor combines bucketing and aggregation
type TimechartProcessor struct {
	stats   *StatsProcessor
	span    time.Duration
	metrics CommandMetrics
}

func (p *TimechartProcessor) Init(ctx context.Context) error { return nil }
func (p *TimechartProcessor) Close() error                 { return nil }
func (p *TimechartProcessor) Metrics() CommandMetrics     { return p.metrics }

func (p *TimechartProcessor) Process(ctx context.Context, in <-chan buffer.Event, out chan<- buffer.Event) error {
	spanSecs := int64(p.span.Seconds())
	if spanSecs <= 0 {
		return fmt.Errorf("invalid timechart span: %v", p.span)
	}

	bucketedIn := make(chan buffer.Event, 1000)
	go func() {
		defer close(bucketedIn)
		for ev := range in {
			p.metrics.EventsIn++
			t, err := time.Parse(time.RFC3339, ev.Timestamp)
			if err == nil {
				t = t.UTC()
				bucketed := (t.Unix() / spanSecs) * spanSecs
				ev.Timestamp = time.Unix(bucketed, 0).UTC().Format(time.RFC3339)
				if ev.Data == nil {
					ev.Data = make(map[string]interface{})
				}
				ev.Data["_time"] = ev.Timestamp
				bucketedIn <- ev
			} else {
				p.metrics.EventsDropped++
			}
		}
	}()

	// Temporary channel to collect stats results and add metadata
	statsOut := make(chan buffer.Event, 1000)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = p.stats.Process(ctx, bucketedIn, statsOut)
		close(statsOut)
	}()

	for ev := range statsOut {
		if ev.Data == nil {
			ev.Data = make(map[string]interface{})
		}
		ev.Data["_span"] = spanSecs
		ev.Data["_type"] = "timechart"
		out <- ev
		p.metrics.EventsOut++
	}
	wg.Wait()
	return nil
}

func ParseSpan(span string) (time.Duration, error) {
	if span == "" {
		return 0, fmt.Errorf("empty span duration")
	}

	// Basic support for s, m, h, d
	multiplier := time.Second
	valStr := span
	if strings.HasSuffix(span, "d") {
		multiplier = 24 * time.Hour
		valStr = strings.TrimSuffix(span, "d")
	} else if strings.HasSuffix(span, "h") {
		multiplier = time.Hour
		valStr = strings.TrimSuffix(span, "h")
	} else if strings.HasSuffix(span, "m") {
		multiplier = time.Minute
		valStr = strings.TrimSuffix(span, "m")
	} else if strings.HasSuffix(span, "s") {
		multiplier = time.Second
		valStr = strings.TrimSuffix(span, "s")
	}

	// If no suffix, try raw ParseDuration
	amt, err := strconv.Atoi(valStr)
	if err == nil {
		d := time.Duration(amt) * multiplier
		if d <= 0 {
			return 0, fmt.Errorf("span must be positive: %s", span)
		}
		return d, nil
	}

	d, err := time.ParseDuration(span)
	if err == nil && d <= 0 {
		return 0, fmt.Errorf("span must be positive: %s", span)
	}
	return d, err
}

// MaskingProcessor applies role-based data redaction
type MaskingProcessor struct {
	masker      *compliance.Masker
	role        string
	metrics     CommandMetrics
	resultHash  [32]byte
	hasHash     bool
}

func (p *MaskingProcessor) Init(ctx context.Context) error { return nil }
func (p *MaskingProcessor) Close() error                 { return nil }
func (p *MaskingProcessor) Metrics() CommandMetrics     { return p.metrics }

func (p *MaskingProcessor) Process(ctx context.Context, in <-chan buffer.Event, out chan<- buffer.Event) error {
	for ev := range in {
		p.metrics.EventsIn++
		ev.Data = p.masker.Mask(ev.Data, p.role)

		// Update cumulative hash for chain of custody
		data, _ := json.Marshal(ev.Data)
		if !p.hasHash {
			p.resultHash = sha256.Sum256(data)
			p.hasHash = true
		} else {
			h := sha256.New()
			h.Write(p.resultHash[:])
			h.Write(data)
			copy(p.resultHash[:], h.Sum(nil))
		}

		out <- ev
		p.metrics.EventsOut++
	}
	return nil
}

func (p *MaskingProcessor) CumulativeHash() string {
	if !p.hasHash {
		return ""
	}
	return hex.EncodeToString(p.resultHash[:])
}
