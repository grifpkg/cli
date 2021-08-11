package config

import "strings"

type Dependency struct {
	Service string
	Resource string
	Author string
}

func ParseResourceString(dependency string) Dependency{
	var author = ""
	var resource = ""
	var service = ""
	if len(dependency)>0 {
		if dependency[0]=='@' {
			var parts = strings.Split(dependency,"/")
			if len(parts)>=2 {
				author = dependency[1:len(parts[0])]
				resource = dependency[1+len(parts[0]):]
			} else {
				resource =dependency
			}
		} else {
			resource =dependency
		}
	}
	return Dependency{service, resource, author}
}