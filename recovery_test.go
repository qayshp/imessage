package main

import (
	"reflect"
	"testing"
	"time"
)

func TestRecoveryPlanRequiresExplicitScope(t *testing.T) {
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	for _, raw := range []string{`{}`, `{"since":"2026-11-06T00:00:00Z","chats":[]}`, `{"since":"2026-12-01T00:00:00Z","chats":["chat"]}`, `{"since":"2026-11-06T00:00:00Z","chats":[""]}`} {
		if _, err := parseRecoveryPlan([]byte(raw), now); err == nil {
			t.Fatalf("accepted invalid plan %s", raw)
		}
	}
	if _, err := parseRecoveryPlan([]byte(`{"since":"2026-11-06T00:00:00Z","chats":["chat"]}`), now); err != nil {
		t.Fatal(err)
	}
}

func TestRecoveryFillsHolesAndMultipartWithoutDuplicates(t *testing.T) {
	// The first and last parts are already recorded; recovery must fill the middle hole.
	known := map[int]bool{0: true, 2: true}
	missing := unrecordedPartIndexes(3, func(part int) bool { return known[part] })
	if !reflect.DeepEqual(missing, []int{1}) {
		t.Fatal(missing)
	}
	for _, part := range missing {
		known[part] = true
	}
	if got := unrecordedPartIndexes(3, func(part int) bool { return known[part] }); len(got) != 0 {
		t.Fatal("second recovery duplicates parts", got)
	}
}

func TestRecoveryDoesNotAffectUnselectedChats(t *testing.T) {
	t.Setenv("MAUTRIX_IMESSAGE_RECOVERY_PLAN", "")
	if _, selected, err := configuredRecoveryStart("chat"); selected || err != nil {
		t.Fatalf("%v %v", selected, err)
	}
}
