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
	attrIndex := make(map[namespacedKey]*XSDAttribute)
	for _, s := range schemas {
		for _, attr := range s.Attributes {
			if attr.Name != "" {
				attrIndex[makeNamespacedKey(s.TargetNamespace, attr.Name)] = attr
			}
		}
	}

	// Next, traverse all attrs with refs and copy over the properties from the referenced attrs.
	var currentSchema *XSDSchema
	attrByRef := func(ref string) *XSDAttribute {
		if currentSchema == nil {
			return nil
		}
		parts := strings.SplitN(ref, ":", 2)
		if len(parts) == 1 {
			return attrIndex[makeNamespacedKey(currentSchema.TargetNamespace, parts[0])]
		} else {
			if ns, ok := currentSchema.Xmlns[parts[0]]; ok {
				return attrIndex[makeNamespacedKey(ns, parts[1])]
			}
		}
		return nil
	}

	(&visitor{all: schemas}).visit(&visitorConfig{
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
					attr.TargetNamespace = currentSchema.XMLNameForAttribute(refAttr).Space
				}
			} else if attr.Type == "" && attr.SimpleType != nil {
				attr.Type = attr.SimpleType.Restriction.Base
			}
		},
	})
}
