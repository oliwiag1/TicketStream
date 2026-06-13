package handlers

import (
	"reflect"
	"testing"
)

func TestNormalizeSeatRows(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "defaults for empty input",
			input: nil,
			want:  []string{"A", "B", "C", "D", "E"},
		},
		{
			name:  "trims uppercases and deduplicates",
			input: []string{" a ", "A", "b", "", "B", " c "},
			want:  []string{"A", "B", "C"},
		},
		{
			name:  "fallback when only blanks",
			input: []string{" ", "  \t"},
			want:  []string{"A", "B", "C", "D", "E"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSeatRows(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("normalizeSeatRows() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeatLockKey(t *testing.T) {
	got := seatLockKey("event-1", "seat-a1")
	want := "seat_lock:event-1:seat-a1"
	if got != want {
		t.Fatalf("seatLockKey() = %q, want %q", got, want)
	}
}
