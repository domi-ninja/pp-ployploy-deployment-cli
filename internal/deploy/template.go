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

func RenderValue(value string, plan Plan, hostID string, env map[string]string) string {
	value = RenderTemplate(value, VarsForPlan(plan))
	replacer := strings.NewReplacer(envReplacements(env)...)
	value = replacer.Replace(value)
	for serviceID, targets := range plan.AutoPorts[hostID] {
		for target, published := range targets {
			value = strings.ReplaceAll(value, "${service."+serviceID+".port."+itoa(target)+"}", itoa(published))
		}
	}
	return value
}

func envReplacements(env map[string]string) []string {
	out := make([]string, 0, len(env)*2)
	for key, value := range env {
		out = append(out, "${env."+key+"}", value)
	}
	return out
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := "0123456789"
	out := ""
	for value > 0 {
		out = string(digits[value%10]) + out
		value /= 10
	}
	return out
}
