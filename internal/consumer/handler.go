package consumer

import "encoding/json"

type HandlerFunc func(json.RawMessage) error
