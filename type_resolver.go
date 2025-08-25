package gowsdl

import (
	"encoding/xml"
	"strings"
)

type typeResolver struct {
	elementsByTypeName map[namespacedName][]elementAndSchema
}

type namespacedName string

func makeQualifiedNameKey(namespace, local string) namespacedName {
	return namespacedName(namespace + "|" + local)
}

type elementAndSchema struct {
	element  *XSDElement
	schema   *XSDSchema
	topLevel bool
}

func newTypeResolver(schemas []*XSDSchema) *typeResolver {
	var currentSchema *XSDSchema
	var currentType *XSDComplexType
	elementsByTypeName := make(map[namespacedName][]elementAndSchema)

	v := &visitor{all: schemas}
	v.visit(&visitorConfig{
		onEnterSchema: func(s *XSDSchema) {
			currentSchema = s
		},
		onEnterComplexType: func(ct *XSDComplexType) {
			currentType = ct
		},
		onExitComplexType: func(ct *XSDComplexType) {
			currentType = nil
		},
		onEnterElement: func(e *XSDElement) {
			if e.Type != "" {
				topLevel := currentType == nil
				parts := strings.SplitN(e.Type, ":", 2)
				var key namespacedName
				if len(parts) == 1 {
					key = makeQualifiedNameKey(
						currentSchema.namespaceForElement(e, topLevel),
						e.Type,
					)
				} else {
					key = makeQualifiedNameKey(currentSchema.Xmlns[parts[0]], parts[1])
				}
				elementsByTypeName[key] = append(elementsByTypeName[key], elementAndSchema{
					element:  e,
					schema:   currentSchema,
					topLevel: topLevel,
				})
			}
		},
	})
	return &typeResolver{
		elementsByTypeName: elementsByTypeName,
	}
}

// Given a type qname and a schema that it appears within, determine the XMLName
// tag that should be used on the Go type representing that XML type.
//
// If no elements with the given type exist return the original typeName plus the schema's
// target namespace.
//
// If all elements with the given type have the same name and namespace, return that name
// and namespace.
//
// Otherwise, if there are mismatched names or namespaces among the elements with the given
// type, return the original typeName plus the schema's target namespace.
//
// Note that this last case will likely result in broken generated code.
// The more correct solution here is probably to not generate XMLName fields at all,
// except for types used directly as request/response wrappers, and instead emit namespaces
// on field tags, but that would be a somewhat involved change.
func (tr *typeResolver) xmlNameForType(typeName string, schema *XSDSchema) xml.Name {
	key := makeQualifiedNameKey(schema.TargetNamespace, typeName)
	elements, ok := tr.elementsByTypeName[key]
	if !ok {
		return xml.Name{
			Space: schema.TargetNamespace,
			Local: typeName,
		}
	}

	elementNames := make(map[string]struct{})
	namespaces := make(map[string]struct{})
	for _, e := range elements {
		elementNames[e.element.Name] = struct{}{}
		namespaces[e.schema.namespaceForElement(e.element, e.topLevel)] = struct{}{}
	}
	if len(elementNames) == 1 && len(namespaces) == 1 {
		return xml.Name{
			Space: elements[0].schema.namespaceForElement(elements[0].element, elements[0].topLevel),
			Local: elements[0].element.Name,
		}
	}

	return xml.Name{
		Space: schema.TargetNamespace,
		Local: typeName,
	}
}
