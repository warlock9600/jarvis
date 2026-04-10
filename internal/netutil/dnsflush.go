package netutil

type FlushStep struct {
	Command string `json:"command"`
	Status  string `json:"status"`
	Output  string `json:"output,omitempty"`
}

type FlushResult struct {
	OS      string      `json:"os"`
	Steps   []FlushStep `json:"steps"`
	Message string      `json:"message"`
}
