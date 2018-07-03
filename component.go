package go_contentline

import "github.com/pkg/errors"

//Component is the outermost structured part and can include multiple other Components and has Properties.
// There are constraints on most Component types concerning which Properties to include and how often. These will not
// be checked here.
type Component struct {
	//Name is the identifying name for this component, e.g. VCARD or VALARM
	// The component identifiers must be iana-registered tokens or have to be prefixed with 'x-'. This will not be checked!
	// The identifier is case-insensitive and will be converted to uppercase when encoding/parsing.
	Name string

	//Properties contains all included Properties.
	Properties []*Property

	//Comps contains all included Components. For vcf-files, this field should be empty (nil), which will not be checked.
	Comps []*Component
}

//Property is the way to include Values into Components. Properties can also have Parameters.
// E.g. the parameter LANG for DESCRIPTION describes the language in which the description is written.
type Property struct {
	//Name is the identifying name for this property, e.g. DESCRIPTION or ROLE
	// The property identifiers must be iana-registered tokens or have to be prefixed with 'x-'. This will not be checked!
	// The identifier is case-insensitive and will be converted to uppercase when encoding/parsing.
	Name string

	//Value is the value for this property, depending on the Name it can have one of multiple types which
	// includes varying restrictions on the format. The Value will be encoded/parsed as-is, meaning without any
	// (un-)escaping of newline characters and so on. Therefore this string must not contain any newline
	// characters (0x0a and 0x0d), as well as any control characters besides HTAB(0x09)
	Value string

	//Parameters contains the property parameters. For details, see below.
	Parameters

	//field for remembering the original form before parsing, see Property.OriginalLine()
	olds string
}

//NewPropertyUnchecked creates a new Property. The property name is checked for validity, see above.
func NewProperty(name, value string, p Parameters) (out *Property, err error) {
	r := ValidID(name)
	if r == nil {
		return NewPropertyUnchecked(name, value, p), nil
	} else {
		return nil, errors.Errorf("Contains at least one illegal character:'%v'", r)
	}
}

//NewPropertyUnchecked creates a new Property, where the property name is not checked for validity
func NewPropertyUnchecked(name, value string, p Parameters) *Property {
	return &Property{name, value, p, ""}
}

//Parameters is a type to represent property parameters as described
// by RFC6868; RFC5545, Section 3.2 and RFC6350, Section 5. The parameter identifiers must
// be iana-registered tokens or have to be prefixed with 'x-'. This will not be checked!
// The identifiers are case-insensitive and will be converted to uppercase when encoding/parsing.
// The parameter values can include any utf8-codepoint, as long as they are not control
// characters (ASCII 0x00 - 0x08,0x0b,0x0c and 0x0e-0x1f), BUT depending on the parameter name a standard could define
// more constraints (e.g. only a defined set of values for VALUE)
type Parameters map[string][]string

//OriginalLine returns the unfolded line from the input, before it was parsed.
// That can be useful for error messages in further conversion into calendar/contact objects.
// This method will return an empty string if this Property was not parsed, but created
func (p *Property) OriginalLine() string {
	return p.olds
}

//AddComponent adds one or more Subcomponents
func (c *Component) AddComponent(subcomps ...*Component) {
	c.Comps = append(c.Comps, subcomps...)
}

//AddProperty adds one or more Properties
func (c *Component) AddProperty(p ...*Property) {
	c.Properties = append(c.Properties, p...)
}

//AddParameter adds one or more Parameter
func (p *Property) AddParameter(key string, val ...string) {
	p.Parameters[key] = append(p.Parameters[key], val...)
}

//find all subcomponents which have the specified name
func (c *Component) FindSubComponents(name string) []*Component {
	var out []*Component = nil
	for _, val := range c.Comps {
		if val.Name == name {
			out = append(out, val)
		}
	}
	return out
}

//find all properties which have the specified name
func (c *Component) FindProperties(name string) []*Property {
	var out []*Property = nil
	for _, val := range c.Properties {
		if val.Name == name {
			out = append(out, val)
		}
	}
	return out
}
