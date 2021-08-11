package client

import (
	"strings"
)

type Dependency struct {
	Service string
	Resource string
	Author string
}


func ParseResourceString(dependency string) Dependency{
	var username = ""
	var name = ""
	var service = ""
	if len(dependency)>0 {
		if dependency[0]=='@' {
			var parts = strings.Split(dependency,"/")
			if len(parts)>=2 {
				username = dependency[1:len(parts[0])]
				name = dependency[1+len(parts[0]):]
			} else {
				name=dependency
			}
		} else {
			name=dependency
		}
	}
	return Dependency{service,name,username}
}
