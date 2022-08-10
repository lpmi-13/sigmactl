package commands

import (
	"fmt"
	"io"

	"github.com/lpmi-13/sigmactl"
	"github.com/lpmi-13/sigmactl/commands/displayers"
	"github.com/lpmi-13/sigmactl/cs"
	"github.com/spf13/viper"
)

// this is the base command configuration
type CmdConfig struct {
	NS      string
	Sigmait sigmactl.Config
	Out     io.Writer
	Args    []string

	initServices       func(*CmdConfig) error
	getContextPassword func() string
	setContextPassword func(string)

	// actual commands per "service"
	Regions func() cs.RegionsService
	Server  func() cs.ServerService
	Balance func() cs.BalanceService
}

// This creates an instance of a CmdConfig
func NewCmdConfig(ns string, sc sigmactl.Config, out io.Writer, args []string, initGocs bool) (*CmdConfig, error) {

	cmdConfig := &CmdConfig{
		NS:      ns,
		Sigmait: sc,
		Out:     out,
		Args:    args,

		initServices: func(c *CmdConfig) error {
			password := c.getContextPassword()
			gocsClient, err := c.Sigmait.GetGocsClient(Trace, password)
			if err != nil {
				return fmt.Errorf("Unable to initialise Cloud Sigma API client: %s", err)
			}
			c.Regions = func() sigma.RegionsService { return sigma.NewRegionsService(gocsClient) }
			c.Server = func() sigma.ServerService { return sigma.NewServerService(gocsClient) }
			c.Balance = func() sigma.BalanceService { return sigma.NewBalanceService(gocsClient) }

			return nil
		},

		getContextPassword: func() string {
			context := Context
			if context == "" {
				context = viper.GetString("context")
			}
			password := ""

			switch context {
			case sigmactl.ArgDefaultContext:
				password = viper.GetString(sigmactl.ArgPassword)
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
			case sigmactl.ArgDefaultContext:
				viper.Set(sigmactl.ArgPassword, password)
			default:
				contexts := viper.GetStringMapString("auth-contexts")
				contexts[context] = password

				viper.Set("auth-contexts", contexts)
			}
		},
	}

	if initGocs {
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

	columnList, err := c.Sigmait.GetStrings(c.NS, sigmactl.ArgFormat)
	if err != nil {
		return err
	}

	withHeaders, err := c.Sigmait.GetBool(c.NS, sigmactl.ArgNoHeader)
	if err != nil {
		return err
	}

	sc.NoHeaders = withHeaders
	sc.ColumnList = columnList
	sc.OutputType = Output

	return sc.Display()
}
