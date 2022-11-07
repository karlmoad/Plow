package common

import (
	"Plow/plow/objects"
	"Plow/plow/utility"
	"errors"
	"strings"
)

type Property struct {
	Name  string
	Value string
	IsKey bool
}

type MetadataObject struct {
	ObjectType int64
	keyIdx     int
	index      map[string]int
	properties []Property
}

type MetadataMap struct {
	objects map[string][]*MetadataObject
}

type Metadata struct {
	contents   map[int64]MetadataMap
	translator objects.ObjectTypeTranslator
}

func NewMetadataObject(objectType int64, properties ...Property) (*MetadataObject, error) {
	keyIdx := findKeyProperty(properties)
	if keyIdx == -1 {
		return nil, errors.New("no key property identified")
	}
	meta := &MetadataObject{ObjectType: objectType, keyIdx: keyIdx, index: make(map[string]int), properties: properties}
	for i, prop := range properties {
		meta.index[strings.ToLower(prop.Name)] = i
	}
	return meta, nil
}

func (mdo *MetadataObject) GetKey() string {
	if mdo.keyIdx > -1 {
		return strings.ToLower(mdo.properties[mdo.keyIdx].Value)
	}
	return ""
}

func (mdo *MetadataObject) Compare(properties ...Property) bool {
	if len(properties) != len(mdo.properties) {
		return false
	}
	passed := 0
	for _, property := range properties {
		if propIdx, ok := mdo.index[strings.ToLower(property.Name)]; ok {
			if strings.Compare(property.Value, mdo.properties[propIdx].Value) == 0 {
				passed += 1
			}
		} else {
			return false
		}
	}
	return passed > 0 && passed == len(mdo.properties)
}

func NewMetadata(translator objects.ObjectTypeTranslator) *Metadata {
	return &Metadata{contents: make(map[int64]MetadataMap), translator: translator}
}

func (m *Metadata) AddObject(obj *MetadataObject) {
	if _, ok := m.contents[obj.ObjectType]; !ok {
		m.contents[obj.ObjectType] = MetadataMap{objects: make(map[string][]*MetadataObject)}
	}
	if _, ok := m.contents[obj.ObjectType].objects[obj.GetKey()]; ok {
		m.contents[obj.ObjectType].objects[obj.GetKey()] = make([]*MetadataObject, 0)
	}

	m.contents[obj.ObjectType].objects[obj.GetKey()] = append(m.contents[obj.ObjectType].objects[obj.GetKey()], obj)
}

func (m *Metadata) FindObjectFromSpec(spec *objects.CodeBlockSpec) (*MetadataObject, error) {
	if m.translator == nil {
		return nil, errors.New("no object type translator provided")
	}

	tipe := m.translator(spec.Type)

	properties := []Property{Property{Name: "name", Value: spec.Object.Name, IsKey: true}}
	if !utility.IsStringEmpty(&spec.Object.Database) {
		properties = append(properties, Property{Name: "database", Value: spec.Object.Database})
	}
	if !utility.IsStringEmpty(&spec.Object.Schema) {
		properties = append(properties, Property{Name: "schema", Value: spec.Object.Schema})
	}

	return m.Find(tipe, properties...)
}

func (m *Metadata) Find(objectType int64, properties ...Property) (*MetadataObject, error) {
	propKeyIdx := findKeyProperty(properties)
	if propKeyIdx == -1 {
		return nil, errors.New("unable to locate key property")
	}
	keyProp := properties[propKeyIdx]

	if level1, ok := m.contents[objectType]; ok {
		if level2, ok := level1.objects[strings.ToLower(keyProp.Value)]; ok {
			for _, obj := range level2 {
				if obj.Compare(properties...) {
					return obj, nil
				}
			}
		}
	}

	return nil, nil
}

func findKeyProperty(properties []Property) int {
	for i, prop := range properties {
		if prop.IsKey {
			return i
		}
	}
	return -1
}
