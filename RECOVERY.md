# Scoped gap recovery

This branch starts at the installed upstream revision `300ba6d0`. Normal startup still fetches from the latest recorded timestamp. To recover older holes, set `MAUTRIX_IMESSAGE_RECOVERY_PLAN` to an operator-owned JSON file:

```json
{"since":"2026-11-05T00:00:00-08:00","chats":["the-exact-source-chat-guid"]}
```

Only the listed chats are queried from `since`. An optional `message_guids` list restricts recovery to those source message GUIDs before attachment downloads; this avoids an unavailable old attachment stalling unrelated gaps. Already recorded message parts and reaction GUIDs are skipped across merged chat IDs. Reactions are processed even when a batch has no new ordinary messages. Repeating a recovery should create no additional events; verify source and bridge GUID coverage after each run. The plan file can contain personal identifiers and should have mode 0600. Remove the environment variable after verification.

No recovery step deletes or edits the message ledger or rewinds timestamps. Take a stopped-bridge database snapshot before deploying a new binary. With bbctl, use `run --custom-startup-command /absolute/path/to/custom-binary --no-override-config sh-imessage` to preserve the custom executable across restarts. The binary must resolve the same libolm used by the installed bridge. Existing config and encryption keys are retained.

Tests: `go test ./...` with the libolm headers/library available. `recovery_test.go` covers invalid/future windows, explicit chat selection and filling a multipart hole without duplicating it on a second run. Live recovery validation also compares source and bridge GUIDs and checks no additional records appear on a second recovery.
