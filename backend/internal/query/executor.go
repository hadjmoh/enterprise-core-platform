package query

import (
	"context"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/storage"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Dispatcher is the high-level entry point for executing SPL queries
type Dispatcher struct {
	storage *storage.FileStorageEngine
}

func NewDispatcher(s *storage.FileStorageEngine) *Dispatcher {
	return &Dispatcher{storage: s}
}

func (d *Dispatcher) Execute(ctx context.Context, spl string) ([]buffer.Event, error) {
	// 1. Lexing & Parsing
	l := NewLexer(spl)
	p := NewParser(l)
	pipeline, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	// 1.5 Extract metadata from last command for visualization intent
	if len(pipeline.Commands) > 0 {
		lastCmd := pipeline.Commands[len(pipeline.Commands)-1]
		pipeline.Meta.Type = lastCmd.Name
		
		// If timechart, extract span
		if lastCmd.Name == "timechart" {
			for _, arg := range lastCmd.Args {
				if strings.HasPrefix(arg.Value, "span=") {
					d, _ := ParseSpan(strings.TrimPrefix(arg.Value, "span="))
					pipeline.Meta.Span = d
				}
			}
		}
	}

	// 2. Compilation (AST -> Processors)
	var processors []Processor
	for _, cmdNode := range pipeline.Commands {
		proc, err := d.compileCommand(cmdNode)
		if err != nil {
			return nil, fmt.Errorf("compilation failed for command %s: %w", cmdNode.Name, err)
		}
		processors = append(processors, proc)
	}

	// 3. Execution
	executor := NewPipelineExecutor(processors, pipeline.Meta)
	return executor.Execute(ctx)
}
func (d *Dispatcher) ExecuteStream(ctx context.Context, spl string) (<-chan buffer.Event, NodeMetadata, <-chan error, error) {
	// 1. Lexing & Parsing
	l := NewLexer(spl)
	p := NewParser(l)
	pipeline, err := p.Parse()
	if err != nil {
		return nil, NodeMetadata{}, nil, fmt.Errorf("parsing failed: %w", err)
	}

	// Extract metadata for visualization intent
	if len(pipeline.Commands) > 0 {
		lastCmd := pipeline.Commands[len(pipeline.Commands)-1]
		pipeline.Meta.Type = lastCmd.Name
		if lastCmd.Name == "timechart" {
			for _, arg := range lastCmd.Args {
				if strings.HasPrefix(arg.Value, "span=") {
					d, _ := ParseSpan(strings.TrimPrefix(arg.Value, "span="))
					pipeline.Meta.Span = d
				}
			}
		}
	}

	// 2. Compilation (AST -> Processors)
	var processors []Processor
	for _, cmdNode := range pipeline.Commands {
		proc, err := d.compileCommand(cmdNode)
		if err != nil {
			return nil, NodeMetadata{}, nil, fmt.Errorf("compilation failed for command %s: %w", cmdNode.Name, err)
		}
		processors = append(processors, proc)
	}

	// 3. Execution (Streaming)
	executor := NewPipelineExecutor(processors, pipeline.Meta)
	results, errs := executor.ExecuteStream(ctx)
	return results, pipeline.Meta, errs, nil
}

func (d *Dispatcher) Explain(spl string) (string, error) {
	l := NewLexer(spl)
	p := NewParser(l)
	pipeline, err := p.Parse()
	if err != nil {
		return "", fmt.Errorf("parsing failed: %w", err)
	}

	explanation := fmt.Sprintf("Query: %s\n", spl)
	explanation += fmt.Sprintf("Pipeline ID: %d\n", pipeline.Meta.Version)
	explanation += "Plan:\n"
	for i, cmd := range pipeline.Commands {
		explanation += fmt.Sprintf("  %d. %s\n", i+1, cmd.Name)
		for _, arg := range cmd.Args {
			if arg.Key != "" {
				explanation += fmt.Sprintf("      - %s = %s\n", arg.Key, arg.Value)
			} else {
				explanation += fmt.Sprintf("      - %s\n", arg.Value)
			}
		}
	}
	return explanation, nil
}

func (d *Dispatcher) compileCommand(node CommandNode) (Processor, error) {
	switch node.Name {
	case "search":
		return NewSearchProcessor(d.storage, node)
	case "where":
		if len(node.Args) == 0 {
			return nil, fmt.Errorf("where command requires a condition")
		}
		proc := &WhereProcessor{}
		for _, arg := range node.Args {
			if arg.Value == "OR" {
				proc.isOr = true
				continue
			}
			pred := Predicate{
				LeftField: arg.Key,
				Operator:  "=", // Default
				RightValue: arg.Value,
			}
			// Check if value is a field reference (this is a heuristic for now)
			if strings.HasPrefix(arg.Value, "$") {
				pred.RightField = strings.TrimPrefix(arg.Value, "$")
			}
			proc.predicates = append(proc.predicates, pred)
		}
		return proc, nil
	case "fields":
		if len(node.Args) == 0 {
			return nil, fmt.Errorf("fields command requires at least one field")
		}
		proc := &FieldsProcessor{fields: make(map[string]bool)}
		for _, arg := range node.Args {
			val := arg.Value
			isKeep := !strings.HasPrefix(val, "-")
			cleanVal := strings.TrimPrefix(strings.TrimPrefix(val, "-"), "+")
			
			// Check for regex
			if strings.HasPrefix(cleanVal, "/") && strings.HasSuffix(cleanVal, "/") {
				reStr := strings.Trim(cleanVal, "/")
				re, err := regexp.Compile(reStr)
				if err != nil {
					return nil, fmt.Errorf("invalid regex in fields: %s", reStr)
				}
				proc.regexes = append(proc.regexes, re)
			} else {
				proc.fields[cleanVal] = true
			}
			proc.isKeep = isKeep
		}
		return proc, nil
	case "eval":
		if len(node.Args) == 0 {
			return nil, fmt.Errorf("eval command requires target=source")
		}
		arg := node.Args[0]
		if arg.Key == "" {
			return nil, fmt.Errorf("invalid eval: %s", arg.Value)
		}
		proc := &EvalProcessor{targetField: arg.Key}
		
		// Heuristic parsing for arithmetic: field+field or field+const
		parts := strings.FieldsFunc(arg.Value, func(r rune) bool {
			return r == '+' || r == '-' || r == '*' || r == '/'
		})
		
		if len(parts) == 1 {
			proc.leftField = arg.Value
		} else if len(parts) == 2 {
			proc.leftField = strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			
			// Find operator
			for _, op := range []string{"+", "-", "*", "/"} {
				if strings.Contains(arg.Value, op) {
					proc.op = op
					break
				}
			}
			
			// Check if right side is constant
			if f, err := strconv.ParseFloat(right, 64); err == nil {
				proc.rightConst = f
				proc.hasConst = true
			} else {
				proc.rightField = right
			}
		}
		return proc, nil
	case "limit":
		if len(node.Args) == 0 {
			return nil, fmt.Errorf("limit command requires numeric argument")
		}
		l, err := strconv.Atoi(node.Args[0].Value)
		if err != nil {
			return nil, fmt.Errorf("invalid limit: %s", node.Args[0].Value)
		}
		return &LimitProcessor{limit: l}, nil
	case "stats":
		if len(node.Args) == 0 {
			return nil, fmt.Errorf("stats command requires aggregation functions")
		}
		var specs []AggregationSpec
		var byFields []string
		inBy := false

		for _, arg := range node.Args {
			if strings.ToLower(arg.Value) == "by" {
				inBy = true
				continue
			}

			if inBy {
				byFields = append(byFields, arg.Value)
			} else {
				val := arg.Value
				alias := arg.Key
				
				// Standard: func(field)
				if strings.Contains(val, "(") && strings.HasSuffix(val, ")") {
					idx := strings.Index(val, "(")
					fn := val[:idx]
					field := val[idx+1:len(val)-1]
					if alias == "" {
						alias = val
					}
					specs = append(specs, AggregationSpec{Func: fn, Field: field, Alias: alias})
				} else {
					// Keyword: count
					if alias == "" {
						alias = val
					}
					specs = append(specs, AggregationSpec{Func: val, Field: "*", Alias: alias})
				}
			}
		}
		return NewStatsProcessor(specs, byFields, MaxGroupsDefault), nil
	case "bucket":
		span := 1 * time.Minute // default
		field := "_time"
		for _, arg := range node.Args {
			if arg.Key == "span" || strings.HasPrefix(arg.Value, "span=") {
				val := arg.Value
				if strings.HasPrefix(val, "span=") {
					val = strings.TrimPrefix(val, "span=")
				}
				d, err := ParseSpan(val)
				if err != nil {
					return nil, err
				}
				span = d
			} else {
				field = arg.Value
			}
		}
		return &BucketProcessor{span: span, field: field}, nil
	case "timechart":
		span := 1 * time.Hour // default
		var specs []AggregationSpec
		var byFields []string
		inBy := false

		for _, arg := range node.Args {
			if strings.HasPrefix(arg.Value, "span=") {
				d, err := ParseSpan(strings.TrimPrefix(arg.Value, "span="))
				if err != nil {
					return nil, err
				}
				span = d
				continue
			}
			if strings.ToLower(arg.Value) == "by" {
				inBy = true
				continue
			}

			if inBy {
				byFields = append(byFields, arg.Value)
			} else {
				// Same function parsing as stats
				val := arg.Value
				alias := arg.Key
				if strings.Contains(val, "(") && strings.HasSuffix(val, ")") {
					idx := strings.Index(val, "(")
					fn := val[:idx]
					field := val[idx+1 : len(val)-1]
					if alias == "" {
						alias = val
					}
					specs = append(specs, AggregationSpec{Func: fn, Field: field, Alias: alias})
				} else {
					if alias == "" {
						alias = val
					}
					specs = append(specs, AggregationSpec{Func: val, Field: "*", Alias: alias})
				}
			}
		}
		// Timechart always groups by _time. We use MaxSeriesDefault for safety.
		byFields = append([]string{"_time"}, byFields...)
		stats := NewStatsProcessor(specs, byFields, MaxSeriesDefault)
		return &TimechartProcessor{stats: stats, span: span}, nil
	default:
		return nil, fmt.Errorf("unsupported command: %s", node.Name)
	}
}
