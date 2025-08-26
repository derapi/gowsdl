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
				attrIndex[newNamespacedKey(s.TargetNamespace, attr.Name)] = attr
			}
		}
	}

	keyFromAttrRef := func(s *XSDSchema, attrRef string) namespacedKey {
		parts := strings.SplitN(attrRef, ":", 2)
		if len(parts) == 1 {
			return newNamespacedKey(s.TargetNamespace, parts[0])
		} else {
			if ns, ok := s.Xmlns[parts[0]]; ok {
				return newNamespacedKey(ns, parts[1])
			}
		}
		return ""
	}

	// Next, traverse all attrs with refs and copy over the properties from the referenced attrs.
	var currentSchema *XSDSchema
	visitor{schemas}.visit(&visitorConfig{
		onEnterSchema: func(s *XSDSchema) {
			currentSchema = s
		},
		onEnterAttribute: func(attr *XSDAttribute) {
			if attr.Ref != "" {
				nsk := keyFromAttrRef(currentSchema, attr.Ref)
				if nsk == "" {
					return
				}
				if refAttr, ok := attrIndex[nsk]; ok && refAttr.Ref == "" {
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
