package main

import (
	"io"
	"log"
	"os"

	"github.com/bromanko/age-plugin-op/plugin"
	"github.com/spf13/cobra"
)

type PluginOptions struct {
	AgePlugin  string
	Generate   bool
	Decrypt    bool
	Encrypt    bool
	OutputFile string
	LogFile    string
}

var example = `
  $ age-plugin-op --generate -o age-identity.txt
  # Created: TODO
  # Recipient: TODO

  [...]

  $ echo "Hello World" | age -r "" > secret.age

  $ age --decrypt -i age-identity.txt -o - secret.age
  Hello World`

var (
	pluginOptions = PluginOptions{}
	rootCmd       = &cobra.Command{
		Use:     "age-plugin-op",
		Long:    "age-plugin-op is a tool to generate age compatible identities backed by 1Password.",
		Example: example,
		RunE:    RunPlugin,
	}
)

func SetLogger() {
	var w io.Writer
	if pluginOptions.LogFile != "" {
		w, _ = os.Open(pluginOptions.LogFile)
	} else if os.Getenv("AGEDEBUG") != "" {
		w = os.Stderr
	} else {
		w = io.Discard
	}
	plugin.SetLogger(w)
}

func RunCli(cmd *cobra.Command) error {
	switch {
	default:
		return cmd.Help()
	}
}

func RunPlugin(cmd *cobra.Command, _ []string) error {
	switch pluginOptions.AgePlugin {
	case "recipient-v1":
		plugin.Log.Println("Got recipient-v1")
	case "identity-v1":
		plugin.Log.Println("Got identity-v1")
	default:
		return RunCli(cmd)
	}
	return nil
}

func pluginFlags(cmd *cobra.Command, _ *PluginOptions) error {
	flags := cmd.Flags()
	flags.SortFlags = false

	//flags.BoolVarP(&pluginOptions.Convert, "convert", "y", false, "Convert identities to recipients.")
	flags.StringVarP(&pluginOptions.OutputFile, "output", "o", "", "Write the result to the file.")

	flags.BoolVarP(&pluginOptions.Generate, "generate", "g", false, "Generate a identity.")
	//flags.BoolVarP(&pluginOptions.PIN, "pin", "p", false, "Include a pin with the key. Alternatively export AGE_TPM_PIN.")

	flags.StringVar(&pluginOptions.LogFile, "log-file", "", "Logging file for debug output")

	//flags.BoolVar(&pluginOptions.SwTPM, "swtpm", false, "Use a software TPM for key storage (Testing only and requires swtpm installed)")

	//flags.BoolVar(&pluginOptions.Decrypt, "decrypt", false, "wip")
	//flags.BoolVar(&pluginOptions.Encrypt, "encrypt", false, "wip")
	flags.StringVar(&pluginOptions.AgePlugin, "age-plugin", "", "internal use")
	//flags.MarkHidden("decrypt")
	//flags.MarkHidden("encrypt")
	return flags.MarkHidden("age-plugin")
}

func main() {
	SetLogger()

	if err := pluginFlags(rootCmd, &pluginOptions); err != nil {
		log.Fatal(err)
	}
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
