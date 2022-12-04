package datasize

import "gopkg.in/yaml.v3"

func (d *DataSize) UnmarshalYAML(node *yaml.Node) error {
	res, err := Parse(node.Value)
	if err != nil {
		return err
	}
	*d = res
	return nil
}

func (d *DataSize) SetValue(s string) error {
	res, err := Parse(s)
	if err != nil {
		return err
	}
	*d = res
	return nil
}

func (d *DataSize) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}
