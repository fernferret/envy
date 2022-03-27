package envy_test

import (
	"os"
	"testing"
	"time"

	"github.com/fernferret/envy"
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
				envy.Disable(tt.args.name)
			}

			if tt.args.custom != "" {
				envy.SetEnvName(tt.args.name, tt.args.custom)
			}

			if tt.args.panic {
				assert.Panics(t, func() { envy.Parse(tt.args.pfx) })
			} else {
				assert.NotPanics(t, func() { envy.Parse(tt.args.pfx) })

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

	assert.Panics(t, func() { envy.Disable("foo") })
}

func TestSetNonexistantFlag(t *testing.T) {
	pflag.CommandLine = pflag.NewFlagSet("test", pflag.PanicOnError)

	pflag.Bool("verbose", false, "test flag")

	assert.Panics(t, func() { envy.SetEnvName("foo", "bar") })
}

func TestDuplicateSet(t *testing.T) {
	pflag.CommandLine = pflag.NewFlagSet("test", pflag.PanicOnError)

	pflag.String("kube-config", "", "test flag")

	assert.NotPanics(t, func() { envy.SetEnvName("kube-config", "KUBECONFIG") })

	// even if it's the same value, panic
	assert.Panics(t, func() { envy.SetEnvName("kube-config", "KUBECONFIG") })
}

func ExampleParse() {
	// Reset CommandLine flags for example, don't include these in your code!
	pflag.CommandLine = pflag.NewFlagSet("test", pflag.PanicOnError)
	os.Clearenv()

	// Define a few flags
	pflag.String("url", "http://localhost:8080", "set the url")
	pflag.Bool("once", false, "only run processing once")
	pflag.DurationP("interval", "i", time.Minute, "interval to check widgets")

	// Simulate COOL_APP_INTERVAL being set
	os.Setenv("COOL_APP_INTERVAL", "10m")

	// Pre-parse with envy with a prefix of COOL_APP_
	envy.Parse("COOL_APP")

	// Don't sort the flags.
	pflag.CommandLine.SortFlags = false

	// Parse the flags
	pflag.Parse()

	// Output results to stdout instead of the default stderr
	pflag.CommandLine.SetOutput(os.Stdout)
	pflag.PrintDefaults()
	// Output: --url string          set the url [COOL_APP_URL] (default "http://localhost:8080")
	//       --once                only run processing once [COOL_APP_ONCE]
	//   -i, --interval duration   interval to check widgets [COOL_APP_INTERVAL 10m] (default 1m0s)
}

func ExampleDisable() {
	// Reset CommandLine flags for example, you don't need this in your code!
	pflag.CommandLine = pflag.NewFlagSet("test", pflag.PanicOnError)

	// Define a few flags
	pflag.String("url", "http://localhost:8080", "set the url")
	pflag.Bool("once", false, "only run processing once")

	// Don't allow FOO_ONCE to set the --once flag.
	envy.Disable("once")

	// Pre-parse with envy with a prefix of FOO_
	envy.Parse("FOO")

	// Parse the flags
	pflag.Parse()

	// Output results to stdout instead of the default stderr
	pflag.CommandLine.SetOutput(os.Stdout)
	pflag.PrintDefaults()
	// Output: --once         only run processing once
	//       --url string   set the url [FOO_URL] (default "http://localhost:8080")
}

func ExampleSetEnvName() {
	// Reset CommandLine flags for example, you don't need this in your code!
	pflag.CommandLine = pflag.NewFlagSet("test", pflag.PanicOnError)

	// Define a few flags
	pflag.String("url", "http://localhost:8080", "set the url")
	pflag.Bool("once", false, "only run processing once")

	// Don't use FOO_URL, instead use MY_HTTP_URL as the env var
	envy.SetEnvName("url", "MY_HTTP_URL")

	// Pre-parse with envy with a prefix of FOO_
	envy.Parse("FOO")

	pflag.Parse()

	// Output results to stdout instead of the default stderr
	pflag.CommandLine.SetOutput(os.Stdout)
	pflag.PrintDefaults()
	// Output: --once         only run processing once [FOO_ONCE]
	//       --url string   set the url [MY_HTTP_URL] (default "http://localhost:8080")
}
