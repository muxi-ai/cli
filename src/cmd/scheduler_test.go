package cmd

import "testing"

func TestSchedulerStatusParts(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{status: "ACTIVE", want: "active"},
		{status: "PAUSED", want: "paused"},
		{status: "COMPLETED", want: "completed"},
		{status: "DISABLED", want: "disabled"},
		{status: "FAILED", want: "failed"},
		{status: "", want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			_, got := schedulerStatusParts(tt.status)
			if got != tt.want {
				t.Errorf("schedulerStatusParts(%q) label = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}
