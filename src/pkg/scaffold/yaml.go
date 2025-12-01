package scaffold

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

// marshalYAML marshals a yaml.Node with 2-space indentation
func marshalYAML(node *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
