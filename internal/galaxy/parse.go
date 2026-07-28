package galaxy

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"go.yaml.in/yaml/v4"
)

// galaxyNameRe matches a namespace.name reference installable from Galaxy.
var galaxyNameRe = regexp.MustCompile(`^[\w-]+\.[\w-]+$`)

// kind distinguishes a collection entry from a role entry.
type kind int

// Entry kinds.
const (
	kindCollection kind = iota
	kindRole
)

// entry is a single collection or role reference within a requirements file.
type entry struct {
	kind    kind
	name    string
	version string
	node    *yaml.Node
	skip    bool
}

type fileData struct {
	path    string
	doc     *yaml.Node
	entries []entry
}

func parseAllFiles(files []string) ([]fileData, error) {
	parsed := make([]fileData, 0, len(files))
	for _, path := range files {
		fd, err := parseFile(path)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, fd)
	}
	return parsed, nil
}

func parseFile(path string) (fileData, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path comes from user args
	if err != nil {
		return fileData{}, fmt.Errorf("read %s: %w", path, err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fileData{}, fmt.Errorf("parse %s: %w", path, err)
	}

	fd := fileData{path: path, doc: &doc}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return fd, nil
	}

	root := doc.Content[0]
	fd.entries = append(fd.entries, parseEntries(mapValue(root, "collections"), kindCollection)...)
	fd.entries = append(fd.entries, parseEntries(mapValue(root, "roles"), kindRole)...)
	return fd, nil
}

// mapValue returns the value node for key in a mapping node, or nil.
func mapValue(node *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func parseEntries(seq *yaml.Node, kind kind) []entry {
	if seq == nil || seq.Kind != yaml.SequenceNode {
		return nil
	}

	entries := make([]entry, 0, len(seq.Content))
	for _, item := range seq.Content {
		//exhaustive:ignore // only scalar and mapping entries are meaningful
		switch item.Kind {
		case yaml.ScalarNode:
			entries = append(entries, entry{
				kind: kind,
				name: item.Value,
				node: item,
				skip: !galaxyNameRe.MatchString(item.Value),
			})
		case yaml.MappingNode:
			entries = append(entries, mappingEntry(item, kind))
		}
	}
	return entries
}

func mappingEntry(item *yaml.Node, kind kind) entry {
	name := scalarValue(item, "name")
	e := entry{
		kind:    kind,
		name:    name,
		version: scalarValue(item, "version"),
		node:    item,
	}

	if !galaxyNameRe.MatchString(name) {
		e.skip = true
		return e
	}
	switch kind {
	case kindCollection:
		e.skip = scalarValue(item, "type") != ""
	case kindRole:
		e.skip = scalarValue(item, "src") != "" || scalarValue(item, "scm") != ""
	}
	return e
}

// scalarValue returns the string value of key in a mapping node, or empty.
func scalarValue(node *yaml.Node, key string) string {
	if v := mapValue(node, key); v != nil {
		return v.Value
	}
	return ""
}

// setVersion sets the entry's version, promoting a scalar entry to a mapping.
func (e *entry) setVersion(version string) {
	item := e.node
	if item.Kind == yaml.ScalarNode {
		m := mappingNode(e.name, version)
		m.HeadComment, m.FootComment = item.HeadComment, item.FootComment
		m.Content[1].LineComment = item.LineComment
		*item = m
		return
	}
	if v := mapValue(item, "version"); v != nil {
		v.Value = version
		return
	}
	item.Content = append(item.Content, scalar("version"), scalar(version))
}

func scalar(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
}

func mappingNode(name, version string) yaml.Node {
	return yaml.Node{
		Kind:    yaml.MappingNode,
		Tag:     "!!map",
		Content: []*yaml.Node{scalar("name"), scalar(name), scalar("version"), scalar(version)},
	}
}

// writeFile re-encodes the document with a leading "---" and 2-space indent.
func writeFile(fd *fileData) error {
	var buf bytes.Buffer
	buf.WriteString("---\n")

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(fd.doc); err != nil {
		return fmt.Errorf("encode %s: %w", fd.path, err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("close encoder %s: %w", fd.path, err)
	}

	if err := os.WriteFile(fd.path, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", fd.path, err)
	}
	return nil
}
