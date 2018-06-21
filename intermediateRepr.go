package go_contentline

import (
	"github.com/pkg/errors"

	"strings"

	"github.com/soh335/icalparser"
)

type Component struct {
	Name       string
	Comps      []*Component
	Properties []*Property
}

func ToIntermediate(obj *icalparser.Object) (out *Component, err error) {
	out = &Component{}
	out.Name = obj.HeaderLine.Value.C
	out.Properties = toProps(obj.PropertiyLines)
	for _, comp := range obj.Components {
		x, e := toIntermediate(comp)
		if e != nil {
			return nil, e
		}
		out.Comps = append(out.Comps, x)
	}
	return
}

func toIntermediate(comp *icalparser.Component) (out *Component, err error) {
	out = &Component{}
	out.Name = comp.HeaderLine.Value.C
	props := comp.PropertiyLines
	for i := 0; i < len(props); i++ {
		if props[i].Name.C == "BEGIN" {
			newcomp := &Component{}
			newcomp.Name = props[i].Value.C
			x, e := parseComponent(props[i+1:], newcomp)
			if e != nil {
				return nil, e
			}
			i = i + x
			out.Comps = append(out.Comps, newcomp)
		} else {
			out.Properties = append(out.Properties, toProp(props[i]))
		}
	}
	return
}

func parseComponent(props []*icalparser.ContentLine, comp *Component) (lines int, err error) {

	for i := 0; i < len(props); i++ {
		if props[i].Name.C == "BEGIN" {
			newcomp := &Component{}
			newcomp.Name = props[i].Value.C
			x, e := parseComponent(props[i+1:], newcomp)
			if e != nil {
				return 0, e
			}
			i = i + x
			comp.Comps = append(comp.Comps, newcomp)
		} else if props[i].Name.C == "END" {
			if props[i].Value.C != comp.Name {
				return 0, errors.New("Unexpected END:" + props[i].Value.C + " , expected END:" + comp.Name)
			} else {
				return i + 1, nil
			}
		} else { //Name = END
			comp.Properties = append(comp.Properties, toProp(props[i]))
		}
	}
	return 0, errors.New("Expected END:" + comp.Name + " , found nothing")
}

type Property struct {
	Name  string
	Value string
	Parameters
	old  *icalparser.ContentLine
	olds string
}
type Parameters map[string][]string

func toProp(line *icalparser.ContentLine) *Property {
	out := &Property{
		Name:       strings.ToLower(line.Name.C),
		Value:      line.Value.C,
		Parameters: make(map[string][]string),
	}
	for _, param := range line.Param {
		for _, val := range param.ParamValues {
			parName := strings.ToLower(param.ParamName.C)
			out.Parameters[parName] = append(out.Parameters[parName], val.C)
		}
	}

	return out
}

func toProps(lines []*icalparser.ContentLine) (out []*Property) {
	out = make([]*Property, len(lines))
	for i, line := range lines {
		out[i] = toProp(line)
	}
	return
}

func (Prop *Property) BeforeParsing() string {
	//old must be the read string representation of this Property ("ContentLine")
	return Prop.olds
}
