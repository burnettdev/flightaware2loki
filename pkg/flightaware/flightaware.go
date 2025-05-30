package flightaware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/burnettdev/flightaware2loki/pkg/alloy"
	"github.com/burnettdev/flightaware2loki/pkg/models"
)

// FetchAndPushToLoki fetches aircraft data from FlightAware and pushes it to Loki
func FetchAndPushToLoki(ctx context.Context, alloyClient *alloy.Client) error {
	// Fetch data from FlightAware
	resp, err := http.Get(os.Getenv("SKYAWARE_URL"))
	if err != nil {
		return fmt.Errorf("failed to fetch FlightAware data: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var data models.AutoGenerated
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode FlightAware data: %w", err)
	}

	// Convert aircraft data to Loki entries
	entries := make([]alloy.LogEntry, 0, len(data.Aircraft))
	for _, aircraft := range data.Aircraft {
		// Create a JSON string of the aircraft data
		aircraftJSON, err := json.Marshal(aircraft)
		if err != nil {
			return fmt.Errorf("failed to marshal aircraft data: %w", err)
		}

		// Create labels for Loki
		labels := map[string]string{
			"service": "flightaware",
		}

		// Create the log entry
		entry := alloy.LogEntry{
			Timestamp: time.Unix(int64(data.Now), 0),
			Labels:    labels,
			Line:      string(aircraftJSON),
		}

		entries = append(entries, entry)
	}

	// Push to Loki
	if err := alloyClient.PushLogs(ctx, entries); err != nil {
		return fmt.Errorf("failed to push logs to Loki: %w", err)
	}

	return nil
}
