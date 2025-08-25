package gowsdl

import (
	"strings"
)

// resolveAttrRefs resolves all attribute references across the provided schemas.
// It modifies the schemas in-place, copying properties from the referenced attributes
// onto the referencing attributes.
func resolveAttrRefs(schemas []*XSDSchema) {
	// First, build an index from (namespace, attrName) -> attrDef
	// for all global attrs across all schemas.
	attrIndex := make(map[string]map[string]*XSDAttribute)
	v := &visitor{all: schemas}
	v.visit(&visitorConfig{
		onEnterSchema: func(s *XSDSchema) {
			attrMap := make(map[string]*XSDAttribute)
			for _, attr := range s.Attributes {
				if attr.Name != "" {
					attrMap[attr.Name] = attr
				}
			}
			attrIndex[s.TargetNamespace] = attrMap
		},
	})

	// Next, traverse all attrs with refs and copy over the properties from the referenced attrs.
	var currentSchema *XSDSchema
	attrByRef := func(ref string) *XSDAttribute {
		if currentSchema == nil {
			return nil
		}
		parts := strings.SplitN(ref, ":", 2)
		if len(parts) == 1 {
			if schemaAttrs, ok := attrIndex[currentSchema.TargetNamespace]; ok {
				return schemaAttrs[parts[0]]
			}
		} else {
			if ns, ok := currentSchema.Xmlns[parts[0]]; ok {
				if schemaAttrs, ok := attrIndex[ns]; ok {
					return schemaAttrs[parts[1]]
				}
			}
		}
		return nil
	}

	v.visit(&visitorConfig{
		onEnterSchema: func(s *XSDSchema) {
			currentSchema = s
		},
		onEnterAttribute: func(attr *XSDAttribute) {
			if attr.Ref != "" {
				refAttr := attrByRef(attr.Ref)
				if refAttr != nil && refAttr.Ref == "" {
					attr.Name = refAttr.Name
					attr.Type = refAttr.Type
					if attr.Fixed == "" {
						attr.Fixed = refAttr.Fixed
					}
				}
			} else if attr.Type == "" && attr.SimpleType != nil {
				attr.Type = attr.SimpleType.Restriction.Base
			}
		},
	})
}
