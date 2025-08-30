package models

type API struct {
	BaseURL   string     `yaml:"base_url"`
	AuthType  string     `yaml:"auth_type"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

type Endpoint struct {
	Path        string            `yaml:"path"`
	Method      string            `yaml:"method"`
	Description string            `yaml:"description"`
	Auth        bool              `yaml:"auth"`
	Request     *EndpointRequest  `yaml:"request,omitempty"`
	Response    *EndpointResponse `yaml:"response,omitempty"`
	Errors      []ErrorResponse   `yaml:"errors,omitempty"`
}

type EndpointRequest struct {
	Params map[string]interface{} `yaml:"params,omitempty"`
	Query  map[string]interface{} `yaml:"query,omitempty"`
	Body   map[string]interface{} `yaml:"body,omitempty"`
}

type EndpointResponse struct {
	Status int                    `yaml:"status"`
	Body   map[string]interface{} `yaml:"body,omitempty"`
}

type ErrorResponse struct {
	Status  int    `yaml:"status"`
	Code    string `yaml:"code"`
	Message string `yaml:"message"`
}
