package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/lpmi-13/sigmactl"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

// ErrUnknownTerminal is when sigmait can't return the current terminal type
// when requesting an auth token
// NOTE: we probably don't need auth tokens since we're doing basic auth
var (
	ErrUnknownTerminal = errors.New("unknown terminal")
	cfgFileWriter      = defaultConfigFileWriter
)

// would be nice to not have to deal with passwords, and use tokens instead, but oh well
func retrieveUserPasswordFromCommandLine() (string, error) {
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		return "", ErrUnknownTerminal
	}

	fmt.Print("Please authenticate sigmactl for use with your cloud sigma account")
	fmt.Print("Enter your account password: ")
	passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	return string(passwordBytes), nil
}

type UnknownSchemeError struct {
	Scheme string
}

var _ error = &UnknownSchemeError{}

func (use *UnknownSchemeError) Error() string {
	return "Unknown scheme: " + use.Scheme
}

// the *very* basic (for now) auth commands
func Auth() *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "auth",
			Short: "Display commands for authenticating with sigmactl",
			Long:  "You need to authenticate with each call before you can use this CLI tool. calling `sigmactl auth init` and passing your credentials will store them locally to be sent with every request so you don't have to keep entering them",
		},
	}

	cmdBuilderWithInit(cmd, RunAuthInit(retrieveUserPasswordFromCommandLine), "init", "Set up your authentication credentials", `This command lets you set up your account credentials locally so you don't have to enter them for every request.`, Writer, false)

}

//RunAuthInit sets up the sigmactl config. This is stored in $XDG_CONFIG_HOME/sigmactl. On Unix, if XDG_CONFIG_HOME is not set, uses $USER/.config. On Windows, uses %APP_DATA%/sigmactl/config.
func RunAuthInit(retrieveUserPasswordFunc func() (string, error)) func(c *CmdConfig) error {

	return func(c *CmdConfig) error {
		password := c.getContextPassword()

		if password == "" {
			in, err := retrieveUserPasswordFunc()
			if err != nil {
				return fmt.Errorf("Unable to read cloud sigma password: %s", err)
			}
			password = strings.TrimSpace(in)
		} else {
			fmt.Fprintf(c.Out, "Using password [%v]", password)
			fmt.Fprintln(c.Out)
		}

		c.setContextPassword(string(password))

		fmt.Fprintln(c.Out)
		fmt.Fprint(c.Out, "Validating password...")

		if err := c.initServices(c); err != nil {
			return fmt.Errorf("Unable to authenticate with this password: %s", err)
		}

		if _, err := c.Account().Get(); err != nil {
			fmt.Fprintln(c.Out, "invalid password")
			fmt.Fprintln(c.Out)
			return fmt.Errorf("Unable to use supplied password to access API: %s", err)
		}

		fmt.Fprintln(c.Out, "OK")
		fmt.Fprintln(c.Out)

		return writeConfig()
	}
}

func writeConfig() error {
	f, err := cfgFileWriter()
	if err != nil {
		return err
	}

	defer f.Close()

	b, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		return errors.New("Unable to write config file.")
	}

	return nil
}

// this write to a config.yaml file in the default config home
func defaultConfigFileWriter() (io.WriteCloser, error) {
	cfgFile := viper.GetString("config")

	defaultCfgFile := filepath.Join(defaultConfigHome(), defaultConfigName)
	if cfgFile == defaultCfgFile {
		configHome()
	}

	f, err := os.Create(cfgFile)
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(cfgFile, 0600); err != nil {
		return nil, err
	}

	return f, nil
}
