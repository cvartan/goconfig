package mapconv

import (
	"maps"
	"strconv"
	"strings"
)

// Преобразование многоуровневого map (который в себе содержит другие map) в плоский map со свойствами
func ParseMapToPropertyMap(source map[string]any) map[string]any {
	return parseMapToPropertyMap("", source)
}

func parseMapToPropertyMap(prefix string, source map[string]any) map[string]any {
	result := make(map[string]any, len(source))

	for k, v := range source {
		key := prefix
		if prefix != "" {
			key = key + "."
		}
		key = key + k

		switch t := v.(type) {
		case map[string]any:
			{
				m := parseMapToPropertyMap(key, t)
				maps.Copy(result, m)
			}
		case []any:
			{
				m := parseArrayToPropertyMap(key, t)
				maps.Copy(result, m)
			}
		default:
			{
				result[key] = t
			}
		}
	}

	return result
}

func parseArrayToPropertyMap(prefix string, source []any) map[string]any {
	result := make(map[string]any, len(source))
	for i, v := range source {
		key := prefix + "." + strconv.Itoa(i)

		switch t := v.(type) {
		case map[string]any:
			{
				m := parseMapToPropertyMap(key, t)
				maps.Copy(result, m)
			}
		case []any:
			{
				m := parseArrayToPropertyMap(key, t)
				maps.Copy(result, m)
			}
		default:
			{
				result[key] = t
			}
		}
	}

	return result
}

// Преобразование плоского map со свойствами в многоуровневый map (который в себе содержит другие map)
func ParsePropertyMapToMap(props map[string]any) map[string]any {
	root := &node{
		ChildNodes: make([]*node, 0, 8),
	}

	for k, v := range props {
		pathValues := strings.Split(k, ".")

		parentNode := root
		var currentNode *node
		for _, val := range pathValues {
			currentNode = parentNode.GetChildByName(val)
			if currentNode == nil {
				currentNode = &node{
					Name:       val,
					ChildNodes: make([]*node, 0, 8),
				}
				parentNode.ChildNodes = append(parentNode.ChildNodes, currentNode)
			}
			parentNode = currentNode
		}
		currentNode.Value = v
	}

	return parsePropertyMapToMap(root)
}

type node struct {
	Name       string
	ChildNodes []*node
	Value      any
}

func (n *node) GetChildByName(name string) *node {
	for _, v := range n.ChildNodes {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func parsePropertyMapToMap(parent *node) map[string]any {
	result := make(map[string]any, len(parent.ChildNodes))
	for _, n := range parent.ChildNodes {
		if len(n.ChildNodes) > 0 {
			result[n.Name] = parsePropertyMapToMap(n)
		} else {
			result[n.Name] = n.Value
		}
	}

	return result
}
