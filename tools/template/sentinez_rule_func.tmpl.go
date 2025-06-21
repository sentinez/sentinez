package template

var SentinezRuleFunc = `
// Code generated. DO NOT EDIT.
package rules

import "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/waf/v1"

var RuleSet = map[string]*waf.Rule{
	{{- range .Rules }}
	{{- $ids := .Actions.Fields.Id }}
	{{- if $ids }}
	"{{ index $ids 0 }}": R{{ index $ids 0 }}(),
	{{- end }}
	{{- end }}
}

{{- range .Rules }}
{{- $rule := . }}
{{- $ids := .Actions.Fields.Id }}
{{- if $ids }}
// R{{ index $ids 0 }} returns rule with ID {{ index $ids 0 }}
func R{{ index $ids 0 }}() *waf.Rule {
	return &waf.Rule{
		Actions: &waf.RuleAction{
			Statement: {{ $rule.Actions.Statement | printf "%q" }},
			{{- if $rule.Actions.Fields }}
			Fields: &waf.RuleActionField{
				{{- if $rule.Actions.Fields.Id }}Id: []string{ {{ range $rule.Actions.Fields.Id }}{{ printf "%q" . }},{{ end }} },{{ end }}
				{{- if $rule.Actions.Fields.Msg }}Msg: []string{ {{ range $rule.Actions.Fields.Msg }}{{ printf "%q" . }},{{ end }} },{{ end }}
				{{- if $rule.Actions.Fields.Phase }}Phase: []string{ {{ range $rule.Actions.Fields.Phase }}{{ printf "%q" . }},{{ end }} },{{ end }}
				{{- if $rule.Actions.Fields.Tag }}Tag: []string{ {{ range $rule.Actions.Fields.Tag }}{{ printf "%q" . }},{{ end }} },{{ end }}
			},
			{{- end }}
		},
		Configuration: {{ $rule.Configuration | printf "%q" }},
		ConfigurationBase64: "{{ $rule.Configuration | base64Encode }}",
	}
}
{{ end -}}
{{ end -}}

`
