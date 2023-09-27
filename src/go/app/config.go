package app

import (
	"os"
	"strings"

	"github.com/synthesio/zconfig"
)

var (
	defaultRepository zconfig.Repository
	defaultProcessor  zconfig.Processor
	argsProvider      = NewArgsProvider()
	envProvider       = zconfig.NewEnvProvider()
)

func init() {
	defaultRepository.AddProviders(argsProvider, envProvider)
	defaultRepository.AddParsers(zconfig.ParseString)
	defaultProcessor.AddHooks(defaultRepository.Hook, zconfig.Initialize)
}

func Configure(s interface{}) error {
	return defaultProcessor.Process(s)
}

func Args() []string {
	return argsProvider.PositionalArgs
}

// ArgsProvider is shamelessly copy-pasted from the zconfig repository with the
// addition of handling PositionArgs.
type ArgsProvider struct {
	Args           map[string]string
	PositionalArgs []string
}

func NewArgsProvider() (p *ArgsProvider) {
	p = new(ArgsProvider)
	p.Args = make(map[string]string, len(os.Args))

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		if !strings.HasPrefix(arg, "--") {
			p.PositionalArgs = append(p.PositionalArgs, arg)
			continue
		}

		arg = strings.TrimPrefix(arg, "--")
		parts := strings.SplitN(arg, "=", 2)

		if len(parts) == 1 && i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "--") {
			parts = append(parts, os.Args[i+1])
			i += 1
		}

		parts = append(parts, "")
		p.Args[parts[0]] = parts[1]
	}

	return p
}

func (p *ArgsProvider) Retrieve(key string) (value interface{}, found bool, err error) {
	value, found = p.Args[key]
	return value, found, nil
}

func (ArgsProvider) Name() string {
	return "args"
}

func (ArgsProvider) Priority() int {
	return 1
}
