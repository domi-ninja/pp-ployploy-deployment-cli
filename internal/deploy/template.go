package deploy

import "strings"

type RenderVars struct {
	Project     string
	Environment string
	GitSHA      string
	ShortSHA    string
	Release     string
}

func VarsForPlan(plan Plan) RenderVars {
	return RenderVars{
		Project:     plan.Config.Project.Name,
		Environment: plan.Config.Project.Environment,
		GitSHA:      plan.Git.SHA,
		ShortSHA:    plan.Git.ShortSHA,
		Release:     plan.ReleaseID,
	}
}

func RenderTemplate(value string, vars RenderVars) string {
	replacer := strings.NewReplacer(
		"${project}", vars.Project,
		"${environment}", vars.Environment,
		"${git_sha}", vars.GitSHA,
		"${short_sha}", vars.ShortSHA,
		"${release}", vars.Release,
	)
	return replacer.Replace(value)
}

func RenderTemplates(values []string, vars RenderVars) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, RenderTemplate(value, vars))
	}
	return out
}
