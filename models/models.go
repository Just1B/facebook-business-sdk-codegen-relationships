package models

type AdObjectApi struct {
	Method   string              `json:"method"`
	Endpoint string              `json:"endpoint"`
	Return   string              `json:"return"`
	Params   []AdObjectApiParams `json:"params"`
}

type AdObjectApiParams struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Type     string `json:"type"`
}

type AdObjectField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type AdObject struct {
	Api    []AdObjectApi   `json:"api"`
	Fields []AdObjectField `json:"fields"`
}
