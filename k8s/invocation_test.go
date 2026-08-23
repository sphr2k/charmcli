package k8s

import "testing"

func TestInvocationDisplayName(t *testing.T) {
	tests := []struct {
		name string
		inv  Invocation
		argv string
		want string
	}{
		{name: "policy standalone", inv: Invocation{Standalone: "policyopsctl", Plugin: "policyops"}, argv: "/usr/local/bin/policyopsctl", want: "policyopsctl"},
		{name: "policy plugin", inv: Invocation{Standalone: "policyopsctl", Plugin: "policyops"}, argv: "/usr/local/bin/kubectl-policyops", want: "kubectl policyops"},
		{name: "resource standalone", inv: Invocation{Standalone: "resourcectl", Plugin: "resources"}, argv: "resourcectl", want: "resourcectl"},
		{name: "resource plugin", inv: Invocation{Standalone: "resourcectl", Plugin: "resources"}, argv: "kubectl-resources", want: "kubectl resources"},
		{name: "other executable", inv: Invocation{Standalone: "resourcectl", Plugin: "resources"}, argv: "/tmp/dev-resource", want: "dev-resource"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.inv.DisplayName(tt.argv); got != tt.want {
				t.Fatalf("DisplayName(%q) = %q, want %q", tt.argv, got, tt.want)
			}
		})
	}
}
