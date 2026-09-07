package models

import "github.com/andreykyz/gostaticstructdiff/examples/models/nested"

// Metadata represents additional key‑value data with mixed types.
type Metadata struct {
	ID     nested.ID         `structtomap:"id"`
	Label  string            `structtomap:"label"`
	Values map[string]string `structtomap:"values"`
	Score  *float64          `structtomap:"score"`
	Extra  struct {
		Note string  `structtomap:"note"`
		Cost float64 `structtomap:"cost"`
	} `structtomap:"extra"`
	MetaTime nested.MetaTime `structtomap:"meta_time"`
}

type MetaMeta map[nested.ID]Metadata
