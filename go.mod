module github.com/crossplane-contrib/provider-gitlab

go 1.16

require (
	github.com/crossplane/crossplane-runtime v0.13.0
	github.com/crossplane/crossplane-tools v0.0.0-20201201125637-9ddc70edfd0d
	github.com/google/go-cmp v0.5.2
	github.com/pkg/errors v0.9.1
	github.com/xanzy/go-gitlab v0.50.0
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43 // indirect
	golang.org/x/tools v0.0.0-20201208062317-e652b2f42cc7 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/controller-tools v0.5.0
)
