package go_contentline

import (
	"io"
	"fmt"
	"strings"
	"unicode/utf8"
)

const maxLineLength = 75

type Component struct {
	Name       string
	Comps      []*Component
	Properties []*Property
}


type Property struct {
	Name  string
	Value string
	Parameters
	olds string
}
type Parameters map[string][]string


func (Prop *Property) BeforeParsing() string {
	//old must be the read string representation of this Property ("ContentLine")
	return Prop.olds
}

func (c *Component) EncodeICal(w io.Writer){
	fmt.Fprintf(w,"%s:%s\r\n",sBEGIN,strings.ToUpper(c.Name))
	for _,p :=range c.Properties{
		p.EncodeICal(w)
	}
	for _,c :=range c.Comps{
		c.EncodeICal(w)
	}

	fmt.Fprintf(w,"%s:%s\r\n",sEND,strings.ToUpper(c.Name))
}

func (p *Property) EncodeICal(w io.Writer){
	out:=strings.ToUpper(p.Name)
	for k,vals:=range p.Parameters{
		out := out + ";" + strings.ToUpper(k) + "="
		for i,v:=range vals{
			if i>0{
				out = out+","
			}
			val:=EscapeParamVal(v)
			if strings.ContainsAny(val,",;:"){
				val = "\"" +val+"\""
			}
			out =out+val
		}
	}
	out = out+":"+p.Value
	parts:=split(out,maxLineLength)
	for i,part:=range parts{
		if i>0{
			fmt.Fprint(w," ")
		}
		fmt.Fprint(w,part)
		fmt.Fprint(w,"\r\n")
	}
}

func split(in string,maxlen int)(out []string){
	if len(in)<=maxlen{
		return []string{in}
	}
	inr:=[]rune(in)
	out = nil
	prev:=0
	sum:=0
	for _,r:=range inr{
		rl:=utf8.RuneLen(r)
		if sum+rl-prev>maxlen{
			out = append(out,in[prev:sum])
			prev=sum
		}
		sum = sum+rl
	}
	return
}


