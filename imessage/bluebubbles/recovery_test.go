package bluebubbles

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRepairSkipsKnownAttachmentBeforeDownloading(t *testing.T) {
	file := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(file, []byte(`{"message_guids":["missing"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MAUTRIX_IMESSAGE_RECOVERY_PLAN", file)
	messages := []Message{{GUID: "known", Attachments: []Attachment{{GUID: "unavailable"}}}, {GUID: "missing"}}
	selected, err := filterRepairMessages(messages, "bridge-repair-test")
	if err != nil || len(selected) != 1 || selected[0].GUID != "missing" {
		t.Fatalf("%v %v", selected, err)
	}
	ordinary, err := filterRepairMessages(messages, "bridge-catchup-test")
	if err != nil || len(ordinary) != 2 {
		t.Fatal("normal catch-up unexpectedly filtered")
	}
}

func TestHTTPTimeoutAndCredentialRedaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	client := &http.Client{Timeout: 20 * time.Millisecond}
	_, err := client.Get(server.URL + "?password=private-test-value")
	if err == nil {
		t.Fatal("expected timeout")
	}
	safe := safeRequestError(err)
	if strings.Contains(safe.Error(), "private-test-value") {
		t.Fatal("credential leaked")
	}
	var timeout interface{ Timeout() bool }
	if !errors.As(safe, &timeout) || !timeout.Timeout() {
		t.Fatal("lost timeout information")
	}
	if apiHTTPClient.Timeout <= 0 || attachmentHTTPClient.Timeout <= 0 {
		t.Fatal("requests must be bounded")
	}
}
