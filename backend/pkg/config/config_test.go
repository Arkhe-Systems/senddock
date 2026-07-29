package config

import "testing"

func TestResolveCloud(t *testing.T) {
	cases := []struct {
		name           string
		cloud          string
		deploymentMode string
		want           bool
	}{
		{name: "unset means self-hosted", want: false},
		{name: "CLOUD=true", cloud: "true", want: true},
		{name: "CLOUD=1", cloud: "1", want: true},
		{name: "CLOUD=yes", cloud: "yes", want: true},
		{name: "CLOUD=TRUE is case insensitive", cloud: "TRUE", want: true},
		{name: "CLOUD=false", cloud: "false", want: false},
		{name: "CLOUD=anything else", cloud: "maybe", want: false},
		{name: "deprecated DEPLOYMENT_MODE still works", deploymentMode: "cloud", want: true},
		{name: "deprecated value is case insensitive", deploymentMode: "Cloud", want: true},
		{name: "deprecated self-hosted stays self-hosted", deploymentMode: "self-hosted", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CLOUD", tc.cloud)
			t.Setenv("DEPLOYMENT_MODE", tc.deploymentMode)

			if got := resolveCloud(); got != tc.want {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestDeploymentModeName(t *testing.T) {
	if got := (Config{IsCloud: true}).DeploymentModeName(); got != "cloud" {
		t.Errorf("expected cloud, got %q", got)
	}
	if got := (Config{}).DeploymentModeName(); got != "self-hosted" {
		t.Errorf("expected self-hosted, got %q", got)
	}
	if (Config{IsCloud: true}).IsSelfHosted() {
		t.Error("a cloud deployment must not report as self-hosted")
	}
}
