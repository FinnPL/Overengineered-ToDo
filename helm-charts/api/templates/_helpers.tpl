{{/*
Expand the name of the chart.
*/}}
{{- define "api.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "api.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version label.
*/}}
{{- define "api.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" -}}
{{- end -}}

{{/*
Normalize a service key for label usage.
*/}}
{{- define "api.serviceKey" -}}
{{- regexReplaceAll "[^a-z0-9-]" (lower .name) "-" -}}
{{- end -}}

{{/*
Compute the display name for a service (without release prefix).
*/}}
{{- define "api.serviceName" -}}
{{- $svc := .values | default dict -}}
{{- if $svc.nameOverride }}
{{- $svc.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- include "api.serviceKey" . -}}
{{- end -}}
{{- end -}}

{{/*
Compute the full resource name for a service.
*/}}
{{- define "api.serviceFullname" -}}
{{- $svc := .values | default dict -}}
{{- if $svc.fullnameOverride }}
{{- $svc.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $namePart := include "api.serviceName" . -}}
{{- printf "%s-%s" (include "api.fullname" .root) $namePart | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Selector labels unique per service.
*/}}
{{- define "api.selectorLabels" -}}
{{- $labels := dict -}}
{{- $labels = set $labels "app.kubernetes.io/name" (include "api.serviceFullname" .) -}}
{{- $labels = set $labels "app.kubernetes.io/instance" .root.Release.Name -}}
{{- $labels = set $labels "app.kubernetes.io/component" (include "api.serviceName" .) -}}
{{- toYaml $labels -}}
{{- end -}}

{{/*
Standard labels applied to service resources before merging user overrides.
*/}}
{{- define "api.serviceStandardLabels" -}}
{{- $root := .root -}}
{{- $labels := dict -}}
{{- $labels = set $labels "helm.sh/chart" (include "api.chart" $root) -}}
{{- $labels = merge $labels (include "api.selectorLabels" . | fromYaml) -}}
{{- $labels = set $labels "app.kubernetes.io/managed-by" $root.Release.Service -}}
{{- if $root.Chart.AppVersion }}
{{- $labels = set $labels "app.kubernetes.io/version" ($root.Chart.AppVersion | toString) -}}
{{- end -}}
{{- toYaml $labels -}}
{{- end -}}

{{/*
Merge standard, global, and service-specific labels.
*/}}
{{- define "api.serviceLabels" -}}
{{- $root := .root -}}
{{- $svc := .values | default dict -}}
{{- $labels := dict -}}
{{- $standard := include "api.serviceStandardLabels" . }}
{{- if $standard }}
{{- $labels = merge $labels ($standard | fromYaml) -}}
{{- end }}
{{- with $root.Values.global.commonLabels }}
{{- $labels = merge $labels (deepCopy .) -}}
{{- end }}
{{- with $svc.labels }}
{{- $labels = merge $labels (deepCopy .) -}}
{{- end }}
{{- toYaml $labels -}}
{{- end -}}

{{/*
Merge annotations with precedence service > global.
*/}}
{{- define "api.mergeAnnotations" -}}
{{- $root := .root -}}
{{- $svc := .values | default dict -}}
{{- $annotations := dict -}}
{{- with $root.Values.global.commonAnnotations }}
{{- $annotations = merge $annotations (deepCopy .) -}}
{{- end }}
{{- with $svc.annotations }}
{{- $annotations = merge $annotations (deepCopy .) -}}
{{- end }}
{{- if gt (len $annotations) 0 }}
{{- toYaml $annotations -}}
{{- end }}
{{- end -}}

{{/*
Merge pod annotations with precedence service > global.
*/}}
{{- define "api.podAnnotations" -}}
{{- $root := .root -}}
{{- $svc := .values | default dict -}}
{{- $annotations := dict -}}
{{- with $root.Values.global.podAnnotations }}
{{- $annotations = merge $annotations (deepCopy .) -}}
{{- end }}
{{- with $svc.podAnnotations }}
{{- $annotations = merge $annotations (deepCopy .) -}}
{{- end }}
{{- if gt (len $annotations) 0 }}
{{- toYaml $annotations -}}
{{- end }}
{{- end -}}

{{/*
Merge pod labels with precedence service > global.
*/}}
{{- define "api.podLabels" -}}
{{- $root := .root -}}
{{- $svc := .values | default dict -}}
{{- $labels := dict -}}
{{- $labels = merge $labels (include "api.selectorLabels" . | fromYaml) -}}
{{- with $root.Values.global.commonLabels }}
{{- $labels = merge $labels (deepCopy .) -}}
{{- end }}
{{- with $svc.podLabels }}
{{- $labels = merge $labels (deepCopy .) -}}
{{- end }}
{{- toYaml $labels -}}
{{- end -}}

{{/*
Determine host for ingress when not explicitly set.
*/}}
{{- define "api.serviceHost" -}}
{{- $root := .root -}}
{{- $svc := .values | default dict -}}
{{- $ingress := $svc.ingress | default dict -}}
{{- if $ingress.host }}
{{- $ingress.host -}}
{{- else -}}
{{- $dns := $root.Values.global.dns | default dict -}}
{{- $tier := default "api" $dns.tierSubdomain -}}
{{- $rootDomain := default "cluster.local" $dns.rootDomain -}}
{{- $env := default "" $root.Values.global.environment -}}
{{- $svcSubdomain := default (include "api.serviceName" .) $ingress.subdomain -}}
{{- if and (default true $dns.includeEnvironmentSubdomain) (ne $env "") -}}
{{- printf "%s.%s.%s.%s" $svcSubdomain $tier $env $rootDomain -}}
{{- else -}}
{{- printf "%s.%s.%s" $svcSubdomain $tier $rootDomain -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Determine which secret should provide the database connection string.
*/}}
{{- define "api.databaseSecretName" -}}
{{- $root := .root -}}
{{- $db := .database | default dict -}}
{{- if $db.existingSecret }}
{{- $db.existingSecret -}}
{{- else if $db.secretName }}
{{- $db.secretName -}}
{{- else -}}
{{- printf "%s-database" (include "api.fullname" $root) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Render the env var for referencing the database connection secret.
*/}}
{{- define "api.databaseEnv" -}}
{{- $root := .root -}}
{{- $svc := .svc | default dict -}}
{{- $db := $svc.database | default $root.Values.database | default dict -}}
{{- $shouldRender := or ($db.url) ($db.existingSecret) ($db.secretName) -}}
{{- if $shouldRender }}
{{- $secretName := include "api.databaseSecretName" (dict "root" $root "database" $db) -}}
{{- $secretKey := $db.secretKey | default "database-url" -}}
{{- $envName := $db.envVarName | default "DATABASE_URL" -}}
- name: {{ $envName }}
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: {{ $secretKey }}
{{- end -}}
{{- end -}}
