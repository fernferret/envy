package envy

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestParseFlagSet(t *testing.T) {
	type exp struct {
		usage string
		value string
	}
	type args struct {
		pfx      string
		name     string
		value    interface{}
		usage    string
		disabled bool
		custom   string
		panic    bool
		env      map[string]string
		exp      exp
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test empty prefix",
			args: args{
				name:  "verbose",
				value: false,
				usage: "verbose usage",
				pfx:   "",
				exp: exp{
					usage: "verbose usage [VERBOSE]",
					value: "false",
				},
			},
		},
		{
			name: "test standard prefix",
			args: args{
				name:  "verbose",
				value: false,
				usage: "verbose usage",
				pfx:   "FOO",
				exp: exp{
					usage: "verbose usage [FOO_VERBOSE]",
					value: "false",
				},
			},
		},
		{
			name: "test lowercase prefix",
			args: args{
				name:  "verbose",
				value: false,
				usage: "verbose usage",
				pfx:   "foo_",
				exp: exp{
					usage: "verbose usage [FOO_VERBOSE]",
					value: "false",
				},
			},
		},
		{
			name: "test env with default",
			args: args{
				name:  "url",
				value: "",
				usage: "url usage",
				env: map[string]string{
					"FOO_URL": "http://127.0.0.1",
				},
				pfx: "FOO",
				exp: exp{
					usage: "url usage [FOO_URL http://127.0.0.1]",
					value: "http://127.0.0.1",
				},
			},
		},
		{
			name: "test bool env with default",
			args: args{
				name:  "verbose",
				value: false,
				usage: "verbose usage",
				env: map[string]string{
					"FOO_VERBOSE": "true",
				},
				pfx: "FOO",
				exp: exp{
					usage: "verbose usage [FOO_VERBOSE true]",
					value: "true",
				},
			},
		},
		{
			name: "test invalid bool env",
			args: args{
				name:  "verbose",
				value: false,
				usage: "verbose usage",
				env: map[string]string{
					"FOO_VERBOSE": "yay",
				},
				panic: true,
				pfx:   "FOO",
			},
		},
		{
			name: "test disabled flag",
			args: args{
				name:     "verbose",
				value:    false,
				disabled: true,
				usage:    "verbose usage",
				env: map[string]string{
					"FOO_VERBOSE": "true",
				},
				pfx: "FOO",
				exp: exp{
					usage: "verbose usage",
					value: "false",
				},
			},
		},
		{
			name: "test custom flag",
			args: args{
				name:   "kube-config",
				value:  "~/.kube/config",
				custom: "KUBECONFIG",
				usage:  "set the kube config",
				pfx:    "BAR",
				exp: exp{
					usage: "set the kube config [KUBECONFIG]",
					value: "~/.kube/config",
				},
			},
		},
		{
			name: "test custom flag with default",
			args: args{
				name:   "kube-config",
				value:  "~/.kube/config",
				custom: "KUBECONFIG",
				usage:  "set the kube config",
				env: map[string]string{
					"KUBECONFIG": "~/.kube/config.ops",
				},
				pfx: "BAR",
				exp: exp{
					usage: "set the kube config [KUBECONFIG ~/.kube/config.ops]",
					value: "~/.kube/config.ops",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for key, val := range tt.args.env {
				os.Setenv(key, val)
			}
			pflag.CommandLine = pflag.NewFlagSet("test", pflag.PanicOnError)

			switch v := tt.args.value.(type) {
			case string:
				pflag.String(tt.args.name, v, tt.args.usage)
			case bool:
				pflag.Bool(tt.args.name, v, tt.args.usage)
			default:
				assert.FailNow(t, "invalid type specified: %T", v)
			}

			if tt.args.disabled {
				Disable(tt.args.name)
			}

			if tt.args.custom != "" {
				SetEnvName(tt.args.name, tt.args.custom)
			}

			if tt.args.panic {
				assert.Panics(t, func() { Parse(tt.args.pfx) })
			} else {
				assert.NotPanics(t, func() { Parse(tt.args.pfx) })

				flag := pflag.Lookup(tt.args.name)

				assert.Equal(t, tt.args.exp.usage, flag.Usage)
				assert.Equal(t, tt.args.exp.value, flag.Value.String())
			}
		})
	}
}

func TestDisableNonexistantFlag(t *testing.T) {
	pflag.CommandLine = pflag.NewFlagSet("test", pflag.PanicOnError)

	pflag.Bool("verbose", false, "test flag")

	assert.Panics(t, func() { Disable("foo") })
}

func TestSetNonexistantFlag(t *testing.T) {
	pflag.CommandLine = pflag.NewFlagSet("test", pflag.PanicOnError)

	pflag.Bool("verbose", false, "test flag")

	assert.Panics(t, func() { SetEnvName("foo", "bar") })
}

func TestDuplicateSet(t *testing.T) {
	pflag.CommandLine = pflag.NewFlagSet("test", pflag.PanicOnError)

	pflag.String("kube-config", "", "test flag")

	assert.NotPanics(t, func() { SetEnvName("kube-config", "KUBECONFIG") })

	// even if it's the same value, panic
	assert.Panics(t, func() { SetEnvName("kube-config", "KUBECONFIG") })
}
