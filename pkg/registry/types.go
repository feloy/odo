package registry

// Registry is the main struct of devfile registry
type Registry struct {
	Name   string `json:"name,omitempty"`
	URL    string `json:"url,omitempty"`
	Secure bool   `json:"secure,omitempty"`
}

// DevfileStack is the main struct for devfile catalog components
type DevfileStack struct {
	Name        string   `json:"name,omitempty"`
	DisplayName string   `json:"display-name,omitempty"`
	Description string   `json:"description,omitempty"`
	Link        string   `json:"link,omitempty"`
	Registry    Registry `json:"registry,omitempty"`
	Language    string   `json:"language,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	ProjectType string   `json:"project-type,omitempty"`
}

// DevfileStackList lists all the Devfile Stacks
type DevfileStackList struct {
	DevfileRegistries []Registry
	Items             []DevfileStack
}

// TypesWithDetails is the list of project types in devfile registries, and their associated devfiles
type TypesWithDetails map[string][]DevfileStack
