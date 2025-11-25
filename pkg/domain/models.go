package domain

type RawSecret struct {
	APIVersion string                 `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                 `json:"kind" yaml:"kind"`
	Metadata   map[string]interface{} `json:"metadata" yaml:"metadata"`
	Data       map[string]string      `json:"data" yaml:"data"`
	Type       string                 `json:"type" yaml:"type"`
}
