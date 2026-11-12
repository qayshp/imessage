package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// A recovery plan is operator supplied and only applies to explicitly selected chats.
// Keeping the source ledger intact preserves event IDs, multipart messages and reactions.
type recoveryPlan struct {
	Since time.Time `json:"since"`
	Chats []string  `json:"chats"`
}

func parseRecoveryPlan(data []byte, now time.Time) (*recoveryPlan, error) {
	var plan recoveryPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, err
	}
	if plan.Since.IsZero() || plan.Since.After(now) || len(plan.Chats) == 0 {
		return nil, fmt.Errorf("since must be a past RFC3339 timestamp and chats must be nonempty")
	}
	for _, chat := range plan.Chats {
		if chat == "" {
			return nil, fmt.Errorf("empty chat identifier")
		}
	}
	return &plan, nil
}

func configuredRecoveryStart(chat string) (time.Time, bool, error) {
	filename := os.Getenv("MAUTRIX_IMESSAGE_RECOVERY_PLAN")
	if filename == "" {
		return time.Time{}, false, nil
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return time.Time{}, false, err
	}
	plan, err := parseRecoveryPlan(data, time.Now())
	if err != nil {
		return time.Time{}, false, err
	}
	for _, selected := range plan.Chats {
		if chat == selected {
			return plan.Since, true, nil
		}
	}
	return time.Time{}, false, nil
}

func unrecordedPartIndexes(count int, exists func(int) bool) []int {
	indexes := make([]int, 0, count)
	for part := 0; part < count; part++ {
		if !exists(part) {
			indexes = append(indexes, part)
		}
	}
	return indexes
}
