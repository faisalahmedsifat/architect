package models

type Project struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	TechStack   struct {
		Backend  string `yaml:"backend"`
		Database string `yaml:"database"`
		Auth     string `yaml:"auth"`
	} `yaml:"tech_stack"`
	BusinessLogic map[string]string `yaml:"business_logic,omitempty"`
}

type ProjectMarkdown struct {
	Content string
}

func (p *Project) ToMarkdown() string {
	md := "# " + p.Name + "\n\n"
	md += "## Overview\n" + p.Description + "\n\n"
	md += "## Tech Stack\n"
	md += "- Backend: " + p.TechStack.Backend + "\n"
	md += "- Database: " + p.TechStack.Database + "\n"
	md += "- Auth: " + p.TechStack.Auth + "\n\n"

	if len(p.BusinessLogic) > 0 {
		md += "## Business Logic\n\n"
		for title, content := range p.BusinessLogic {
			md += "### " + title + "\n"
			md += content + "\n\n"
		}
	}

	return md
}
