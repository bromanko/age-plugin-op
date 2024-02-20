package main

import (
	"bufio"
	"encoding/base64"
	"filippo.io/age"
	"github.com/bromanko/age-plugin-op/plugin"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"strings"
	"time"
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
  # Created: 2024-02-16 13:25:00.433868 -0800 PST m=+0.003075709
  # Recipient: age1op102xjaf99y9u69cf64cl8trptuenerd3gal8t4hc2exd8z4ntvpyquwaf9l

  AGE-PLUGIN-OP-1Q9D7XC8RDFW0X3F9P7R9WGZDTST5V22CQUMUM3MK6VTKWKJ[...]

  $ echo "Hello World" | age -r "age1op102xjaf99y9u69cf64cl8trptuenerd3gal8t4hc2exd8z4ntvpyquwaf9l" > secret.age

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
		w, _ = os.OpenFile(pluginOptions.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else if os.Getenv("AGEDEBUG") != "" {
		w = os.Stderr
	} else {
		w = io.Discard
	}
	plugin.SetLogger(w)
}

func RunCli(cmd *cobra.Command, out io.Writer) error {
	switch {
	case pluginOptions.Generate:
		if pluginOptions.OutputFile != "" && pluginOptions.OutputFile != "-" {
			f, err := os.OpenFile(pluginOptions.OutputFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
			if err != nil {
				return err
			}
			defer f.Close()
			out = f
		}

		id, err := plugin.CreateIdentity("/Users/bromanko/Code/age-plugin-op/id_ed25519")
		if err != nil {
			return err
		}
		if err = plugin.MarshalIdentity(id, out); err != nil {
			return err
		}
	default:
		return cmd.Help()
	}
	return nil
}

func b64Decode(s string) ([]byte, error) {
	return base64.RawStdEncoding.Strict().DecodeString(s)
}

func b64Encode(s []byte) string {
	return base64.RawStdEncoding.Strict().EncodeToString(s)
}

func RunRecipientV1(stdin io.Reader, stdout io.StringWriter) error {
	time.Sleep(10 * time.Second) // Allow time to attach debugger todo remove

	var entry string
	scanner := bufio.NewScanner(stdin)

	var recipients []string
	var identities []string
	var fileKeys []string
	// Phase 1
parser:
	for scanner.Scan() {
		entry = scanner.Text()
		if len(entry) == 0 {
			continue
		}
		entry = strings.TrimPrefix(entry, "-> ")
		cmd := strings.SplitN(entry, " ", 2)
		plugin.Log.Printf("scanned: '%s'\n", cmd[0])

		switch cmd[0] {
		case "add-recipient":
			plugin.Log.Printf("add-recipient: %s\n", cmd[1])
			recipients = append(recipients, cmd[1])
		case "add-identity":
			plugin.Log.Printf("add-identity: %s\n", cmd[1])
			identities = append(identities, cmd[1])
		case "wrap-file-key":
			scanner.Scan()
			keyB64 := scanner.Text()
			plugin.Log.Printf("wrap-file-key: %s\n", keyB64)
			fileKeys = append(fileKeys, keyB64)
		case "done":
			break parser
		}
	}

	// Phase 2
	var stanzas []*age.Stanza
	var errors []*age.Stanza
	var opRecipients []*plugin.OpRecipient
	for i, recipient := range recipients {
		r, err := plugin.DecodeRecipient(recipient)
		if err != nil {
			plugin.Log.Println("failed to decode recipient: %w", err)
			errors = append(errors, plugin.NewIndexedErrorStanza("recipient", i, err))
		}
		opRecipients = append(opRecipients, r)
	}
	for _, identity := range identities {
		i := plugin.ParseIdentity(identity)
		opRecipients = append(opRecipients, i.Recipient())
	}
	for _, fileKeyb64 := range fileKeys {
		for _, recipient := range opRecipients {
			fileKey, err := b64Decode(fileKeyb64)
			if err != nil {
				return err
			}
			s, err := recipient.Wrap(fileKey)
			if err != nil {
				plugin.Log.Println("failed to wrap file key: %w", err)
				errors = append(errors, plugin.NewInternalErrorStanza(err))
			}
			stanzas = append(stanzas, s...)
		}
	}

	// todo write the stanzas and the errors

	_, _ = stdout.WriteString("-> done\n\n")
	return nil
}

func RunPlugin(cmd *cobra.Command, _ []string) error {
	switch pluginOptions.AgePlugin {
	case "recipient-v1":
		plugin.Log.Println("Got recipient-v1")
		if err := RunRecipientV1(os.Stdin, os.Stdout); err != nil {
			_, _ = os.Stdout.WriteString("-> error\n")
			_, _ = os.Stdout.WriteString(b64Encode([]byte(err.Error())) + "\n")
			return err
		}
	case "identity-v1":
		plugin.Log.Println("Got identity-v1")
	default:
		return RunCli(cmd, os.Stdout)
	}
	return nil
}

func pluginFlags(cmd *cobra.Command, _ *PluginOptions) error {
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVarP(&pluginOptions.OutputFile, "output", "o", "", "Write the result to the file.")

	flags.BoolVarP(&pluginOptions.Generate, "generate", "g", false, "Generate a identity.")

	flags.StringVar(&pluginOptions.LogFile, "log-file", "", "Logging file for debug output")

	flags.StringVar(&pluginOptions.AgePlugin, "age-plugin", "", "internal use")
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
