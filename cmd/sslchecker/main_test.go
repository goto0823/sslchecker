package main

import (
	"testing"
	"time"
)

func TestDaysUntil(t *testing.T) {
	now := time.Date(2026, 9, 3, 17, 15, 0, 0, time.UTC)
	tests := []struct {
		name string
		d    time.Duration
		want int
	}{
		{"30日", 24 * 30 * time.Hour, 30},
		{"30日12時間の端数あり", 24*30*time.Hour + 12*time.Hour, 30},
		{"残り0日", 0, 0},
		{"12時間すぎていた", -12 * time.Hour, -1},
		{"12時間残ってる", 12 * time.Hour, 0},
		{"24時間すぎていた", -24 * time.Hour, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notAfter := now.Add(tt.d)
			got := daysUntil(notAfter, now)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
