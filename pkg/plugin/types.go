package plugin

type ReductMode string

const (
	ModeLabels  ReductMode = "labels"
	ModeContent ReductMode = "content"
	ModeBoth    ReductMode = "both"
)

type reductOptions struct {
	Start      int64      `json:"start,omitempty"`
	Stop       int64      `json:"stop,omitempty"`
	When       any        `json:"when,omitempty"`
	Strict     bool       `json:"strict,omitempty"`
	Continuous bool       `json:"continuous,omitempty"`
	Ext        any        `json:"ext,omitempty"`
	Mode       ReductMode `json:"mode,omitempty"`
}

type reductQuery struct {
	Bucket  string        `json:"bucket"`
	Entry   string        `json:"entry"`
	Options reductOptions `json:"options"`
}
