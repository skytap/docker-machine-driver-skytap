package api

const (
	TemplatePath = "templates"
)

/*
 Skytap template resource.
 */
type Template struct {
	Id              string       `json:"id"`
	Url            string        `json:"url"`
	Name            string       `json:"name"`
	Region          string       `json:"region"`
}
