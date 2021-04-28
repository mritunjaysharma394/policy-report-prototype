module policyreport

go 1.16

require (
	github.com/mritunjaysharma394/policy-report-prototype v0.0.0
	k8s.io/api v0.20.5
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v0.20.5
	k8s.io/code-generator v0.20.1
	sigs.k8s.io/controller-runtime v0.8.3
)

replace github.com/mritunjaysharma394/policy-report-prototype v0.0.0 => ./
