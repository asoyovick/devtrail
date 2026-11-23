package insight

type Insight struct {
	Type           string `json:"type"`
	Scope          string `json:"scope"`
	Summary        string `json:"summary"`
	Observation    string `json:"observation"`
	Evidence       any    `json:"evidence"`
	Recommendation string `json:"recommendation,omitempty"`
	Confidence     string `json:"confidence"`
}
