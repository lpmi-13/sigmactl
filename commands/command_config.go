package commands

import (
	"fmt"
	"io"

	"github.com/lpmi-13/sigmacli"
	"github.com/lpmi-13/sigmacli/commands/displayers"
	"github.com/lpmi-13/sigmacli/do"
	"github.com/spf13/viper"
)

// this is the base command configuration
type CmdConfig struct {
	NS      string
	Sigmait sigmacli.Config
	Out     io.Writer
	Args    []string

	initServices       func(*CmdConfig) error
	getContextPassword func() string
	setContextPassword func(string)

	// actual commands per "service"
	Regions func() sigma.RegionsService
	Server  func() sigma.ServerService
	Balance func() sigma.BalanceService
}

// This creates an instance of a CmdConfig
func NewCmdConfig(ns string, sc sigmacli.Config, out io.Writer, args []string, initGodo bool) (*CmdConfig, error) {

	cmdConfig := &CmdConfig{
		NS:      ns,
		Sigmait: sc,
		Out:     out,
		Args:    args,

		initServices: func(c *CmdConfig) error {
			password := c.getContextPassword()
			godoClient, err := c.Sigmait.GetGodoClient(Trace, password)
			if err != nil {
				return fmt.Errorf("Unable to initialise Cloud Sigma API client: %s", err)
			}
			c.Regions = func() sigma.RegionsService { return sigma.NewRegionsService(godoClient) }
			c.Server = func() sigma.ServerService { return sigma.NewServerService(godoClient) }
			c.Balance = func() sigma.BalanceService { return sigma.NewBalanceService(godoClient) }

			return nil
		},

		getContextPassword: func() string {
			context := Context
			if context == "" {
				context = viper.GetString("context")
			}
			password := ""

			switch context {
			case sigmacli.ArgDefaultContext:
				password = viper.GetString(sigmacli.ArgPassword)
			default:
				contexts := viper.GetStringMapString("auth-contexts")

				password = contexts[context]
			}

			return password

		},

		setContextPassword: func(password string) {
			context := Context
			if context == "" {
				context = viper.GetString("context")
			}

			switch context {
			case sigmacli.ArgDefaultContext:
				viper.Set(sigmacli.ArgPassword, password)
			default:
				contexts := viper.GetStringMapString("auth-contexts")
				contexts[context] = password

				viper.Set("auth-contexts", contexts)
			}
		},
	}

	if initGodo {
		if err := cmdConfig.initServices(cmdConfig); err != nil {
			return nil, err
		}
	}

	return cmdConfig, nil
}

// runs a command and passes in a cmdConfig
type CmdRunner func(*CmdConfig) error

// displays the output of a command
func (c *CmdConfig) Display(d displayers.Displayable) error {
	sc := &displayers.Displayer{
		Item: d,
		Out:  c.Out,
	}

	columnList, err := c.Sigmait.GetStrings(c.NS, sigmacli.ArgFormat)
	if err != nil {
		return err
	}

	withHeaders, err := c.Sigmait.GetBool(c.NS, sigmacli.ArgNoHeader)
	if err != nil {
		return err
	}

	sc.NoHeaders = withHeaders
	sc.ColumnList = columnList
	sc.OutputType = Output

	return sc.Display()
}
