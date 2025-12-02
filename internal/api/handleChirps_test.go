package api

import (
	"testing"
	"time"
)

func TestSortChirpsByCreatedAt(t *testing.T) {
	// Create test chirps with different CreatedAt times
	now := time.Now()
	chirp1 := CompleteChirp{
		ID:        "1",
		CreatedAt: now.Add(-2 * time.Hour), // oldest
		Body:      "First chirp",
		UserID:    "user1",
	}
	chirp2 := CompleteChirp{
		ID:        "2",
		CreatedAt: now.Add(-1 * time.Hour), // middle
		Body:      "Second chirp",
		UserID:    "user2",
	}
	chirp3 := CompleteChirp{
		ID:        "3",
		CreatedAt: now, // newest
		Body:      "Third chirp",
		UserID:    "user3",
	}

	tests := []struct {
		name      string
		chirps    []CompleteChirp
		sortOrder string
		wantOrder []string // expected IDs in order
	}{
		{
			name:      "descending order - newest first",
			chirps:    []CompleteChirp{chirp1, chirp2, chirp3},
			sortOrder: "desc",
			wantOrder: []string{"3", "2", "1"}, // newest to oldest
		},
		{
			name:      "ascending order - oldest first",
			chirps:    []CompleteChirp{chirp3, chirp1, chirp2},
			sortOrder: "asc",
			wantOrder: []string{"1", "2", "3"}, // oldest to newest
		},
		{
			name:      "default order (invalid sortOrder) - should be ascending",
			chirps:    []CompleteChirp{chirp3, chirp1, chirp2},
			sortOrder: "invalid",
			wantOrder: []string{"1", "2", "3"}, // should default to ascending
		},
		{
			name:      "empty sortOrder - should be ascending",
			chirps:    []CompleteChirp{chirp3, chirp1, chirp2},
			sortOrder: "",
			wantOrder: []string{"1", "2", "3"}, // should default to ascending
		},
		{
			name:      "single chirp - desc",
			chirps:    []CompleteChirp{chirp1},
			sortOrder: "desc",
			wantOrder: []string{"1"},
		},
		{
			name:      "single chirp - asc",
			chirps:    []CompleteChirp{chirp1},
			sortOrder: "asc",
			wantOrder: []string{"1"},
		},
		{
			name:      "empty slice",
			chirps:    []CompleteChirp{},
			sortOrder: "desc",
			wantOrder: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of the slice to avoid modifying the original
			chirpsCopy := make([]CompleteChirp, len(tt.chirps))
			copy(chirpsCopy, tt.chirps)

			// Sort the chirps
			sortChirpsByCreatedAt(chirpsCopy, tt.sortOrder)

			// Verify the order
			if len(chirpsCopy) != len(tt.wantOrder) {
				t.Fatalf("Expected %d chirps, got %d", len(tt.wantOrder), len(chirpsCopy))
			}

			for i, wantID := range tt.wantOrder {
				if chirpsCopy[i].ID != wantID {
					t.Errorf("Position %d: expected ID %s, got %s", i, wantID, chirpsCopy[i].ID)
				}
			}

			// Verify that times are actually in the correct order
			if len(chirpsCopy) > 1 {
				if tt.sortOrder == "desc" {
					// Verify descending: each chirp should be newer than the next
					for i := 0; i < len(chirpsCopy)-1; i++ {
						if chirpsCopy[i].CreatedAt.Before(chirpsCopy[i+1].CreatedAt) {
							t.Errorf("Descending order violated at position %d: %v is before %v",
								i, chirpsCopy[i].CreatedAt, chirpsCopy[i+1].CreatedAt)
						}
					}
				} else {
					// Verify ascending: each chirp should be older than the next
					for i := 0; i < len(chirpsCopy)-1; i++ {
						if chirpsCopy[i].CreatedAt.After(chirpsCopy[i+1].CreatedAt) {
							t.Errorf("Ascending order violated at position %d: %v is after %v",
								i, chirpsCopy[i].CreatedAt, chirpsCopy[i+1].CreatedAt)
						}
					}
				}
			}
		})
	}
}

