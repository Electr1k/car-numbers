package job

import (
	"encoding/json"
	"fmt"
	"time"
)

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var raw string

	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("duration must be a string: %w", err)
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("parse duration %q: %w", raw, err)
	}

	*d = Duration(parsed)

	return nil
}
