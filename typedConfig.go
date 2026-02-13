package confection

import (
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

type node struct {
	node  *yaml.Node
	_type string
}

func (n *node) UnmarshalYAML(value *yaml.Node) error {
	n.node = &yaml.Node{
		Kind:        value.Kind,
		Style:       value.Style,
		Tag:         value.Tag,
		Value:       value.Value,
		Anchor:      value.Anchor,
		Alias:       value.Alias,
		Line:        value.Line,
		Column:      value.Column,
		HeadComment: value.HeadComment,
		LineComment: value.LineComment,
		FootComment: value.FootComment,
	}

	typePosition := -1
	for i, content := range value.Content {
		if content.Value == "@type" {
			typePosition = i
			break
		}
	}
	if typePosition == -1 {
		return fmt.Errorf("line %d: @type not found in typed_config", value.Line)
	}
	if typePosition+1 >= len(value.Content) {
		return fmt.Errorf("line %d: @type has no value in typed_config", value.Content[typePosition].Line)
	}
	n._type = value.Content[typePosition+1].Value

	n.node.Content = slices.Concat(value.Content[:typePosition], value.Content[typePosition+2:])

	return nil
}

type TypedConfig struct {
	Name        string     `yaml:"name"`
	TypedConfig *yaml.Node `yaml:"typed_config"`
	_type       string
	line        int
	column      int
}

func (c *TypedConfig) String() string {
	return fmt.Sprintf("Name: %s, TypedConfig: %v, Type: %s", c.Name, c.TypedConfig, c._type)
}

func (c *TypedConfig) Type() string {
	return c._type
}

func (c *TypedConfig) UnmarshalYAML(value *yaml.Node) error {
	type T struct {
		Name        string `yaml:"name"`
		TypedConfig *node  `yaml:"typed_config"`
	}
	var t T
	if err := value.Decode(&t); err != nil {
		return err
	}

	if t.TypedConfig == nil {
		return fmt.Errorf("line %d: typed_config is required", value.Line)
	}

	c.Name = t.Name
	c._type = t.TypedConfig._type
	c.TypedConfig = t.TypedConfig.node
	c.line = value.Line
	c.column = value.Column

	return nil
}
