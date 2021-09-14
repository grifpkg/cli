package project

type DependencyIdentifier struct {
	Version string	`json:"version"`
	Resource string	`json:"resource"`
	Release string 	`json:"release"`
	Hash []string	`json:"hash"`
}