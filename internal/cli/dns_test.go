package cli

import (
	"testing"

	"github.com/lu-zhengda/netwhiz/internal/network"
)

func TestMatchesDNSPreset(t *testing.T) {
	tests := []struct {
		name    string
		servers []string
		preset  []string
		want    bool
	}{
		{
			name:    "exact match",
			servers: []string{"1.1.1.1", "1.0.0.1"},
			preset:  []string{"1.1.1.1", "1.0.0.1"},
			want:    true,
		},
		{
			name:    "different order",
			servers: []string{"1.0.0.1", "1.1.1.1"},
			preset:  []string{"1.1.1.1", "1.0.0.1"},
			want:    false,
		},
		{
			name:    "different length",
			servers: []string{"1.1.1.1"},
			preset:  []string{"1.1.1.1", "1.0.0.1"},
			want:    false,
		},
		{
			name:    "empty servers",
			servers: nil,
			preset:  []string{"1.1.1.1"},
			want:    false,
		},
		{
			name:    "both empty",
			servers: nil,
			preset:  nil,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesDNSPreset(tt.servers, tt.preset)
			if got != tt.want {
				t.Errorf("matchesDNSPreset(%v, %v) = %v, want %v", tt.servers, tt.preset, got, tt.want)
			}
		})
	}
}

func TestBenchmarkDNS(t *testing.T) {
	// Test that benchmarkDNS returns a result (even if the DNS is unreachable in CI).
	result := benchmarkDNS("Test", "127.0.0.1")
	if result.Provider != "Test" {
		t.Errorf("expected provider=Test, got %s", result.Provider)
	}
	if result.Server != "127.0.0.1" {
		t.Errorf("expected server=127.0.0.1, got %s", result.Server)
	}
	// Result may be reachable or not depending on environment.
}

func TestRunDNSBenchmark(t *testing.T) {
	report := runDNSBenchmark([]string{"1.1.1.1"})

	if report == nil {
		t.Fatal("expected non-nil report")
	}

	if len(report.Results) == 0 {
		t.Error("expected at least one result")
	}

	if report.TestedAt.IsZero() {
		t.Error("expected non-zero TestedAt")
	}

	// Verify all standard providers are included.
	providers := make(map[string]bool)
	for _, r := range report.Results {
		providers[r.Server] = true
	}

	for _, expected := range []string{"1.1.1.1", "8.8.8.8", "9.9.9.9"} {
		if !providers[expected] {
			t.Errorf("expected provider with server %s in results", expected)
		}
	}
}

func TestRunDNSBenchmark_NoDuplicateCurrentDNS(t *testing.T) {
	// When current DNS is already one of the preset servers, it should not be duplicated.
	report := runDNSBenchmark([]string{"1.1.1.1"})

	count := 0
	for _, r := range report.Results {
		if r.Server == "1.1.1.1" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("expected 1.1.1.1 to appear exactly once, got %d times", count)
	}
}

func TestRunDNSBenchmark_WithCustomCurrentDNS(t *testing.T) {
	// When current DNS is a custom server, it should be added.
	report := runDNSBenchmark([]string{"10.0.0.1"})

	found := false
	for _, r := range report.Results {
		if r.Server == "10.0.0.1" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected custom current DNS 10.0.0.1 in results")
	}
}

func TestSynthesizeSummary(t *testing.T) {
	tests := []struct {
		name   string
		checks []network.DiagnoseCheck
		want   string
	}{
		{
			name: "all ok",
			checks: []network.DiagnoseCheck{
				{Status: "ok"},
				{Status: "ok"},
			},
			want: "All checks passed. Network is healthy.",
		},
		{
			name: "with warnings",
			checks: []network.DiagnoseCheck{
				{Status: "ok"},
				{Status: "warning"},
			},
			want: "Found 1 warning(s). Network is functional but may have issues.",
		},
		{
			name: "with errors",
			checks: []network.DiagnoseCheck{
				{Status: "error"},
				{Status: "warning"},
			},
			want: "Found 1 error(s) and 1 warning(s). Network connectivity issues detected.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := synthesizeSummary(tt.checks)
			if got != tt.want {
				t.Errorf("synthesizeSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}
