package gowsdl

type typeResolver struct {
	elementNamesByTypeName map[string]map[string]bool
}

func newTypeResolver(schemas []*XSDSchema) *typeResolver {
	elementNamesByTypeName := make(map[string]map[string]bool)

	v := &visitor{all: schemas}
	v.visit(&visitorConfig{
		onEnterElement: func(e *XSDElement) {
			if e.Type != "" {
				typ := stripns(e.Type)
				if _, ok := elementNamesByTypeName[typ]; !ok {
					elementNamesByTypeName[typ] = make(map[string]bool)
				}
				elementNamesByTypeName[typ][e.Name] = true
			}
		},
	})
	return &typeResolver{
		elementNamesByTypeName: elementNamesByTypeName,
	}
}

// Given a type, check if there is an Element with that type, and return its name.
// If multiple elements with identical names of the given type are found,
// the name is returned.
// If multiple elements with different names of the given type are found,
// the original type name is returned instead.
// If no elements are found, the original type name is returned instead.
func (tr *typeResolver) elementNameForType(typeName string) string {
	elementNames, ok := tr.elementNamesByTypeName[typeName]
	if !ok {
		return typeName
	}

	if len(elementNames) == 1 {
		for en := range elementNames {
			return en
		}
	}

	return typeName
}
