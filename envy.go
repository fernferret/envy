package envy

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

const (
	// Used to disable envy entirely for the given flag.
	envyDisable = "envy_disable"

	// Used to set an override that ignores the prefix, useful for well known
	// environment variables like KUBECONFIG
	envyCustom = "envy_custom"
)

var (
	ErrFlagNotExists        = errors.New("flag does not exist")
	ErrCustomAlreadyDefined = errors.New("custom flag already exists")
	ErrInvalidBoolFlagValue = errors.New("bool flag got value that was't 'true' or 'false'")
)

// Parse will call the ParseFlagSet on the default pflag.CommandLine. This is
// likely the function you want to use, but for more details on usage see
// ParseFlagSet.
func Parse(pfx string) {
	ParseFlagSet(pfx, pflag.CommandLine)
}

// ParseFlagSet will loop through defined flags and automatically add an environment
// variable parser for the flag name. This Parse func must be called before the
// call to pflag.Parse() and after you've defined all your flags.
func ParseFlagSet(pfx string, fs *pflag.FlagSet) {

	// Transform the pfx to uppercase and remove trailing _s, this allows many
	// different uses without producing weird results
	if pfx != "" {
		pfx = strings.TrimSuffix(strings.ToUpper(pfx), "_") + "_"
	}

	fs.VisitAll(func(f *pflag.Flag) {

		// Skip any items with envyDisable set at all, there's no way to set it
		// as "false"
		if _, ok := f.Annotations[envyDisable]; ok {
			return
		}

		var envName string
		if val, ok := f.Annotations[envyCustom]; ok {
			// Envy will panic if duplicate custom overrides are defined, so
			// this is always safe to pull the first item.
			envName = val[0]
		} else {
			envName = fmt.Sprintf("%s%s", pfx, strings.ReplaceAll(strings.ToUpper(f.Name), "-", "_"))
		}

		envUsage := envName
		if val, ok := os.LookupEnv(envName); ok {
			envUsage = fmt.Sprintf("%s %s", envName, val)

			// Bool flags are a bit more interesting. I don't want to silently
			// fail if someone passes "yes", so let's panic to blow this thing
			// wide open!
			switch f.Value.Type() {
			case "bool":
				if _, err := strconv.ParseBool(val); err != nil {
					panic(ErrInvalidBoolFlagValue)
				}
			}

			// We can always set this value since the parse function will always
			// win and override us.
			f.Value.Set(val)
		}

		f.Usage = fmt.Sprintf("%s [%s]", f.Usage, envUsage)
	})
}

func Disable(name string) {
	DisableOnFlagSet(name, pflag.CommandLine)
}

func DisableOnFlagSet(name string, fs *pflag.FlagSet) {
	f := fs.Lookup(name)
	if f == nil {
		panic(ErrFlagNotExists)
	}
	if f.Annotations == nil {
		f.Annotations = make(map[string][]string)
	}
	f.Annotations[envyDisable] = []string{"true"}
}

func SetEnvName(name, envName string) {
	SetEnvNameOnFlagSet(name, envName, pflag.CommandLine)
}

func SetEnvNameOnFlagSet(name, envName string, fs *pflag.FlagSet) {
	f := fs.Lookup(name)
	if f == nil {
		panic(ErrFlagNotExists)
	}
	if f.Annotations == nil {
		f.Annotations = make(map[string][]string)
	} else if _, ok := f.Annotations[envyCustom]; ok {
		// Only allow one to be defined, this will prevent weird errors related
		// to copying an envy line and forgetting to change the first flag.
		panic(ErrCustomAlreadyDefined)
	}
	envName = strings.ToUpper(strings.ReplaceAll(envName, "-", "_"))
	f.Annotations[envyCustom] = []string{envName}
}
