package models

// Metadata represents additional key‑value data with mixed types.
type MetadataDiff struct {
	Label struct {
		Value string
		Set   bool
	}
	Values struct {
		Add    map[string]string
		Delete map[string]string
		Set    bool
	}
	Score struct {
		Value *float64
		Set   bool
	}
	Extra struct {
		Value struct {
			Note struct {
				Value string
				Set   bool
			}
			Cost struct {
				Value float64
				Set   bool
			}
		}
		Set bool
	}
}
