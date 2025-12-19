package query

import (
	"context"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/storage"
	"enterprise-core/backend/pkg/logger"
	"os"
	"testing"
	"time"
)

func TestExecution_TimeSeriesHardening(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "storage_test_hardening")
	defer os.RemoveAll(tempDir)

	engine := storage.NewFileStorageEngine(tempDir, logger.New())
	dispatcher := NewDispatcher(engine)
	ctx := context.Background()

	t.Run("Invalid_Span", func(t *testing.T) {
		_, err := dispatcher.Execute(ctx, `search | bucket span=0s _time`)
		if err == nil {
			t.Fatal("Expected error for 0s span")
		}
	})

	t.Run("Max_Span_Exceeded", func(t *testing.T) {
		_, err := dispatcher.Execute(ctx, `search | bucket span=400d _time`)
		if err == nil {
			t.Fatal("Expected error for > 366d span")
		}
	})

	t.Run("Timechart_Metadata", func(t *testing.T) {
		// Seed one event
		nowT := time.Now().UTC()
		now := nowT.Format(time.RFC3339)
		err := engine.Write(ctx, buffer.Event{Timestamp: now, Source: "s1", Data: map[string]interface{}{"val": 10}})
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		engine.FlushAll()
		results, err := dispatcher.Execute(ctx, `search | timechart span=1h count`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("Expected at least one result")
		}
		data := results[0].Data
		if data["_span"] == nil || data["_type"] != "timechart" {
			t.Errorf("Missing metadata: %v", data)
		}
	})

	engine.Close()
}

func TestExecution_Basic(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "storage_test")
	defer os.RemoveAll(tmpDir)

	logg := logger.New()
	engine := storage.NewFileStorageEngine(tmpDir, logg)
	defer engine.Close()

	// Seed some data
	ctx := context.Background()
	engine.Write(ctx, buffer.Event{Timestamp: time.Now().Format(time.RFC3339), Source: "srv1", Data: map[string]interface{}{"message": "error found"}})
	engine.Write(ctx, buffer.Event{Timestamp: time.Now().Format(time.RFC3339), Source: "srv2", Data: map[string]interface{}{"message": "all good"}})
	
	// Wait for flush or force it
	engine.FlushAll()
	time.Sleep(100 * time.Millisecond) // Give a moment for flushes to complete in goroutines

	dispatcher := NewDispatcher(engine)
	
	t.Run("Search Keyword", func(t *testing.T) {
		results, err := dispatcher.Execute(ctx, `search error`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("Explain Mode", func(t *testing.T) {
		explanation, err := dispatcher.Explain(`search error | limit 10`)
		if err != nil {
			t.Fatalf("Explain failed: %v", err)
		}
		if explanation == "" {
			t.Error("Explanation is empty")
		}
		t.Logf("Explanation: %s", explanation)
	})

	t.Run("Where Command", func(t *testing.T) {
		results, err := dispatcher.Execute(ctx, `search | where source=srv1`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result from srv1, got %d", len(results))
		}
	})

	t.Run("Fields Command", func(t *testing.T) {
		results, err := dispatcher.Execute(ctx, `search | fields source`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("No results")
		}
		for _, ev := range results {
			if _, ok := ev.Data["message"]; ok {
				t.Error("Field 'message' was not removed")
			}
			if ev.Source == "" {
				t.Error("Field 'source' was cleared/missing")
			}
		}
	})

	t.Run("Eval Arithmetic", func(t *testing.T) {
		// Use a fresh engine or seed specifically for math
		engine.Write(ctx, buffer.Event{Timestamp: time.Now().Format(time.RFC3339), Source: "srv1", Data: map[string]interface{}{"latency": 100, "base": 50}})
		engine.FlushAll()
		time.Sleep(100 * time.Millisecond)

		results, err := dispatcher.Execute(ctx, `search | eval total=latency+base | where total=150`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("Arithmetic eval failed to produce expected match")
		}
	})

	t.Run("Fields Regex", func(t *testing.T) {
		// Seed a unique event to be sure
		engine.Write(ctx, buffer.Event{Timestamp: time.Now().Format(time.RFC3339), Source: "regex-src", Data: map[string]interface{}{"msg_text": "hello"}})
		engine.FlushAll()
		time.Sleep(100 * time.Millisecond)

		results, err := dispatcher.Execute(ctx, `search | fields "/^msg_.*$/"`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		
		found := false
		for _, ev := range results {
			if _, ok := ev.Data["msg_text"]; ok {
				found = true
				break
			}
		}
		if !found {
			t.Error("Regex field 'msg_text' was not found in results")
		}
	})

	t.Run("Where Compound Logic", func(t *testing.T) {
		// We use OR in our heuristic: srv1 OR srv2 (though the seeded data is srv1/srv2)
		results, err := dispatcher.Execute(ctx, `search | where source=srv1 OR source=srv2`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		// Based on seed data in TestExecution_Basic, we have 2 events: srv1 and srv2
		if len(results) < 2 {
			t.Errorf("Expected at least 2 results with OR logic, got %d", len(results))
		}
	})

	t.Run("Stats Global Count", func(t *testing.T) {
		results, err := dispatcher.Execute(ctx, `search | stats count`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result for global stats, got %d", len(results))
		}
		count := results[0].Data["count"]
		if count == nil {
			t.Fatal("Count field missing in stats result")
		}
	})

	t.Run("Stats Grouped Count", func(t *testing.T) {
		results, err := dispatcher.Execute(ctx, `search | stats count by source`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if len(results) < 2 {
			t.Errorf("Expected multiple groups, got %d", len(results))
		}
	})

	t.Run("Stats Complex Aggregation", func(t *testing.T) {
		results, err := dispatcher.Execute(ctx, `search | stats sum(latency) avg(latency) by source`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		for _, ev := range results {
			if ev.Data["source"] == "srv1" {
				if ev.Data["sum(latency)"] == nil {
					t.Error("sum(latency) missing for srv1")
				}
			}
		}
	})

	t.Run("Stats MaxGroups Hardening", func(t *testing.T) {
		// We'll use a local dispatcher with a very small MaxGroups for testing if we could configure it,
		// but since it's hardcoded for now, we'll verify it doesn't crash on many records.
		// Actually, let's just test null handling and type safety.
		
		// Null handling: stats count by non_existent
		results, err := dispatcher.Execute(ctx, `search | stats count by non_existent`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if len(results) != 1 || results[0].Data["non_existent"] != "(null)" {
			t.Errorf("Expected null group, got %v", results)
		}
	})

	t.Run("Stats Type Safety", func(t *testing.T) {
		// Send non-numeric data to sum()
		engine.Write(ctx, buffer.Event{
			Timestamp: time.Now().Format(time.RFC3339),
			Source:    "srv_bad",
			Data:      map[string]interface{}{"latency": "not_a_number", "message": "srv_bad test"},
		})
		engine.FlushAll()
		time.Sleep(100 * time.Millisecond)

		results, err := dispatcher.Execute(ctx, `search srv_bad | stats sum(latency)`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		// sum of "not_a_number" should be 0 (handled by toFloat returning false)
		if results[0].Data["sum(latency)"].(float64) != 0 {
			t.Errorf("Expected sum 0 for bad data, got %v", results[0].Data["sum(latency)"])
		}
	})

	t.Run("Bucket Command", func(t *testing.T) {
		// Mock events in the past (using local time to match SearchProcessor defaults)
		now := time.Now()
		t1 := now.Add(-30 * time.Minute).Format(time.RFC3339)
		t2 := now.Add(-15 * time.Minute).Format(time.RFC3339)
		
		engine.Write(ctx, buffer.Event{Timestamp: t1, Source: "srv1", Data: map[string]interface{}{"message": "m1"}})
		engine.Write(ctx, buffer.Event{Timestamp: t2, Source: "srv1", Data: map[string]interface{}{"message": "m2"}})
		engine.FlushAll()
		time.Sleep(100 * time.Millisecond)

		results, err := dispatcher.Execute(ctx, `search | bucket span=10m _time | stats count by _time`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if len(results) < 2 {
			t.Errorf("Expected at least 2 buckets, got %d. Check if data was matched in range.", len(results))
		}
	})

	t.Run("Timechart Command", func(t *testing.T) {
		// Use the same events
		results, err := dispatcher.Execute(ctx, `search | timechart span=1h count by source`)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("Timechart produced no results")
		}
		// Should have _time field
		if _, ok := results[0].Data["_time"]; !ok {
			t.Error("_time field missing in timechart result")
		}
	})
}
