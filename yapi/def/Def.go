package def

type Swagger map[string]interface{}

type Header struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	In          string `json:"in"`
	Required    bool   `json:"required"`
}
