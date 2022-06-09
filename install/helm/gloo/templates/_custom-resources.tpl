{{/* vim: set filetype=mustache: */}}

{{/*
Render Gloo Edge custom resources. This is done in a helper function, rather than included as a template,
to ensure that we don't try to apply the custom resources before the service backing the validation webhook
is ready.
*/}}
{{- define "gloo.customResources" -}}
{{- include "gloo.customResources.defaultGateways" . }}
{{ end }}