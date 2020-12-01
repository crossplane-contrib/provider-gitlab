module github.com/crossplane-contrib/provider-gitlab

go 1.13

require (
	cloud.google.com/go/logging v1.1.2
	github.com/crossplane/crossplane-runtime v0.11.0
	github.com/crossplane/crossplane-tools v0.0.0-20201026195708-a544f360b8ac
	github.com/go-ini/ini v1.62.0
	github.com/hashicorp/go-retryablehttp v0.6.4
	github.com/pkg/errors v0.9.1
	github.com/xanzy/go-gitlab v0.39.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.2.4
)
