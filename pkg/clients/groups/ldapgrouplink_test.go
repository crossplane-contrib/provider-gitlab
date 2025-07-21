package groups

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
)

func TestGenerateAddLdapGroupLinkOptions(t *testing.T) {
	cn := "ldap-cn"
	groupAccess := 10
	ldapProvider := "ldapmain"

	type args struct {
		parameters *v1alpha1.LdapGroupLinkParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.AddGroupLDAPLinkOptions
	}{
		"AllFields": {
			args: args{
				parameters: &v1alpha1.LdapGroupLinkParameters{
					CN:           cn,
					GroupAccess:  v1alpha1.AccessLevelValue(groupAccess),
					LdapProvider: ldapProvider,
				},
			},
			want: &gitlab.AddGroupLDAPLinkOptions{
				CN:          &cn,
				GroupAccess: (*gitlab.AccessLevelValue)(&groupAccess),
				Provider:    &ldapProvider,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateAddLdapGroupLinkOptions(tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
