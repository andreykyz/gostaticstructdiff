package models

import "github.com/andreykyz/gostaticstructdiff/examples/models/nested"

// Metadata represents additional key‑value data with mixed types.
type Metadata struct {
	Label  string            `structtomap:"label"`
	Values map[string]string `structtomap:"values"`
	Score  *float64          `structtomap:"score"`
	Extra  struct {
		Note string  `structtomap:"note"`
		Cost float64 `structtomap:"cost"`
	} `structtomap:"extra"`
}

type Metadatas map[nested.ID]Metadata
