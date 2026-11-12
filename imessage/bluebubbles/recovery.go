package bluebubbles

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var apiHTTPClient = &http.Client{Timeout: 30 * time.Second}
var attachmentHTTPClient = &http.Client{Timeout: 2 * time.Minute}

func safeRequestError(err error) error {
	if wrapped, ok := err.(*url.Error); ok {
		return wrapped.Err
	}
	return err
}

// Filter before conversion: already-recorded attachments may no longer be downloadable.
func filterRepairMessages(messages []Message, backfillID string) ([]Message, error) {
	filename := os.Getenv("MAUTRIX_IMESSAGE_RECOVERY_PLAN")
	if filename == "" || !strings.HasPrefix(backfillID, "bridge-repair-") {
		return messages, nil
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var plan struct {
		GUIDs []string `json:"message_guids"`
	}
	if err = json.Unmarshal(data, &plan); err != nil {
		return nil, err
	}
	if len(plan.GUIDs) == 0 {
		return messages, nil
	}
	wanted := make(map[string]bool, len(plan.GUIDs))
	for _, guid := range plan.GUIDs {
		wanted[guid] = true
	}
	selected := make([]Message, 0, len(messages))
	for _, message := range messages {
		if wanted[message.GUID] {
			selected = append(selected, message)
		}
	}
	return selected, nil
}
