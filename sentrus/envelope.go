package sentrus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

const envelopeSeparators = 3

type (
	envelopeHeader struct {
		EventID string    `json:"event_id"`
		SentAt  time.Time `json:"sent_at"`
	}

	itemHeader struct {
		ContentType string `json:"content_type"`
		Length      int    `json:"length"`
		Type        string `json:"type"`
	}
)

func buildEnvelope(event event) ([]byte, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("encoding event: %w", err)
	}

	headerData, err := json.Marshal(envelopeHeader{EventID: event.EventID, SentAt: time.Now().UTC()})
	if err != nil {
		return nil, fmt.Errorf("encoding envelope header: %w", err)
	}

	itemData, err := json.Marshal(itemHeader{
		ContentType: "application/json",
		Length:      len(eventData),
		Type:        "event",
	})
	if err != nil {
		return nil, fmt.Errorf("encoding item header: %w", err)
	}

	var envelope bytes.Buffer
	envelope.Grow(len(headerData) + len(itemData) + len(eventData) + envelopeSeparators)
	envelope.Write(headerData)
	envelope.WriteByte('\n')
	envelope.Write(itemData)
	envelope.WriteByte('\n')
	envelope.Write(eventData)
	envelope.WriteByte('\n')
	return envelope.Bytes(), nil
}
