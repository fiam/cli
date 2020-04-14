package command

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)

	configGetCmd.Flags().StringP("host", "h", "", "Set per-host setting")
	configSetCmd.Flags().StringP("host", "h", "", "Set per-host setting")

}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set and get gh settings",
	Long: `
	TODO
`,
}

// TODO use cobra validators for arg length

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "TODO",
	RunE:  configGet,
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "TODO",
	RunE:  configSet,
}

func configGet(cmd *cobra.Command, args []string) error {
	// TODO rely on validator

	if len(args) != 1 {
		return errors.New("need to pass just a key")
	}
	key := args[0]
	hostname, err := cmd.Flags().GetString("host")
	if err != nil {
		return err
	}

	ctx := contextForCommand(cmd)

	cfg, err := ctx.Config()
	if err != nil {
		return err
	}

	val, err := cfg.Get(hostname, key)
	if err != nil {
		return err
	}

	if val != "" {
		fmt.Println(val)
	}

	return nil
}

func configSet(cmd *cobra.Command, args []string) error {
	// TODO rely on validator
	if len(args) != 2 {
		return errors.New("need to pass a key and a value")
	}

	key := args[0]
	value := args[1]

	hostname, err := cmd.Flags().GetString("host")
	if err != nil {
		return err
	}

	ctx := contextForCommand(cmd)

	cfg, err := ctx.Config()
	if err != nil {
		return err
	}

	err = cfg.Set(hostname, key, value)
	if err != nil {
		return fmt.Errorf("failed to set %q to %q: %s", key, value, err)
	}

	err = cfg.Write()
	if err != nil {
		return fmt.Errorf("failed to write config to disk: %s", err)
	}

	return nil
}
