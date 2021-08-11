package config

type Project struct {
	Name         string		 								`json:"name"`
	Type         string      								`json:"type"`
	Private		 bool										`json:"private"`
	Version		 string										`json:"version"`
	InstallPaths map[string]string 							`json:"installPaths"`
	Dependencies map[string]DependencyIdentifier 			`json:"dependencies"`
	ExcludeFiles []string									`json:"excludeFiles"`
}

type DependencyIdentifier struct {
	Version string	`json:"version"`
	Resource string	`json:"resource"`
	Release string 	`json:"release"`
	Hash []string	`json:"hash"`
}

func getDefaultInstallPaths() map[string]string{
	return map[string]string {
		"default":	"./plugins/",
		"sk":		"./plugins/Skript/scripts/",
	}
}

func GetDefaultProject() Project {
	return Project{"","minecraft/bukkit",true,"1.0.0",getDefaultInstallPaths(),map[string]DependencyIdentifier{},[]string{}}
}