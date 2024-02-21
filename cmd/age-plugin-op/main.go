package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"filippo.io/age"
	"fmt"
	"github.com/bromanko/age-plugin-op/plugin"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type PluginOptions struct {
	AgePlugin  string
	Generate   string
	OutputFile string
	LogFile    string
}

var example = `
  $ age-plugin-op --generate "op://Personal/wxrzetxonuggniebjzruxycq/private key" -o age-identity.txt
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
		Long:    "age-plugin-op is a tool to generate age compatible identities backed by 1Password SSH keys.",
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
	case pluginOptions.Generate != "":
		if pluginOptions.OutputFile != "" && pluginOptions.OutputFile != "-" {
			f, err := os.OpenFile(pluginOptions.OutputFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
			if err != nil {
				return err
			}
			defer f.Close()
			out = f
		}

		id, err := plugin.CreateIdentity(pluginOptions.Generate)
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

func b64Encode(s []byte) string {
	return base64.RawStdEncoding.Strict().EncodeToString(s)
}

func b64Decode(s string) ([]byte, error) {
	return base64.RawStdEncoding.Strict().DecodeString(s)
}

func respondWithStanzas(w io.Writer, errors, stanzas []*age.Stanza) error {
	if len(errors) > 0 {
		for _, e := range errors {
			err := plugin.MarshalStanza(e, w)
			if err != nil {
				return err
			}
		}
	} else {
		for _, s := range stanzas {
			err := plugin.MarshalStanza(s, w)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func RunRecipientV1(stdin io.Reader, stdout io.Writer) error {
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
		for i, recipient := range opRecipients {
			fileKey, err := b64Decode(fileKeyb64)
			if err != nil {
				errors = append(errors, plugin.NewInternalErrorStanza(err))
			}
			wrapStanzas, err := recipient.Wrap(fileKey)
			if err != nil {
				plugin.Log.Println("failed to wrap file key: %w", err)
				errors = append(errors, plugin.NewInternalErrorStanza(err))
			}
			for _, wrapStanza := range wrapStanzas {
				tag := b64Encode(recipient.Tag())
				s := &age.Stanza{
					Type: "recipient-stanza",
					Args: append([]string{strconv.Itoa(i), wrapStanza.Type, tag}, wrapStanza.Args...),
					Body: wrapStanza.Body,
				}
				stanzas = append(stanzas, s)
			}
		}
	}

	err := respondWithStanzas(stdout, errors, stanzas)
	if err != nil {
		return err
	}

	_, _ = io.WriteString(stdout, "-> done\n\n")
	return nil
}

var footerPrefix = []byte("---")

func RunIdentityV1(stdin io.Reader, stdout io.Writer) error {
	var recipients []*age.Stanza
	var identities []*age.Stanza
	var fileKeys []*age.Stanza

	//var entry string
	rr := bufio.NewReader(stdin)
	sr := plugin.NewStanzaReader(rr)

	// Phase 1
parser:
	for {
		peek, err := rr.Peek(len(footerPrefix))
		if err != nil {
			return fmt.Errorf("failed to read header: %w", err)
		}

		if bytes.Equal(peek, footerPrefix) {
			line, err := rr.ReadBytes('\n')
			if err != nil {
				return fmt.Errorf("failed to read header: %w", err)
			}

			prefix, args := plugin.SplitArgs(line)
			if prefix != string(footerPrefix) || len(args) != 1 {
				return fmt.Errorf("malformed closing line: %q", line)
			}
			mac, err := plugin.DecodeString(args[0])
			if err != nil || len(mac) != 32 {
				return fmt.Errorf("malformed closing line %q: %v", line, err)
			}
			break
		}

		s, err := sr.ReadStanza()
		if err != nil {
			return fmt.Errorf("failed to parse header: %w", err)
		}

		switch s.Type {
		case "add-recipient":
			plugin.Log.Printf("add-recipient: %s\n", s.Args)
			recipients = append(recipients, s)
		case "add-identity":
			plugin.Log.Printf("add-identity: %s\n", s.Args)
			identities = append(identities, s)
		case "recipient-stanza":
			plugin.Log.Printf("recipient-stanza: %s\n", s.Args)
			fileKeys = append(fileKeys, s)
		case "done":
			break parser
		}
	}

	// Phase 2
	var stanzas []*age.Stanza
	var errors []*age.Stanza
	var opIdentities []*plugin.OpIdentity
	for i, recipient := range recipients {
		r, err := plugin.DecodeRecipient(recipient.Args[0])
		if err != nil {
			plugin.Log.Println("failed to decode recipient: %w", err)
			errors = append(errors, plugin.NewIndexedErrorStanza("recipient", i, err))
		}
		opIdentities = append(opIdentities, r.Identity())
	}
	for i, identity := range identities {
		identity, err := plugin.DecodeIdentity(identity.Args[0])
		if err != nil {
			plugin.Log.Println("failed to decode identity: %w", err)
			errors = append(errors, plugin.NewIndexedErrorStanza("identity", i, err))
			continue
		}
		opIdentities = append(opIdentities, identity)
	}
	for i, fileKey := range fileKeys {
		if fileKey.Args[1] != "ssh-rsa" && fileKey.Args[1] != "ssh-ed25519" {
			plugin.Log.Println("not an ssh key")
			continue
		}

		_, err := b64Decode(fileKey.Args[2])
		if err != nil {
			return fmt.Errorf("failed base64 decode tag: %v", err)
		}

		_, err = b64Decode(fileKey.Args[3])
		if err != nil {
			return fmt.Errorf("failed base64 decode session key: %v", err)
		}

		// find the identity with the matching tag
		var matchingIdentity *plugin.OpIdentity
		for _, identity := range opIdentities {
			tag, _ := b64Decode(fileKey.Args[2])
			if bytes.Equal(identity.Recipient().Tag(), tag) {
				matchingIdentity = identity
			}
		}
		if matchingIdentity == nil {
			return fmt.Errorf("no matching identity found for tag: %v", fileKey.Args[2])
		}

		sshStanza := &age.Stanza{
			Type: fileKey.Args[1],
			Args: fileKey.Args[3:],
			Body: fileKey.Body,
		}
		unwrappedKey, err := matchingIdentity.Unwrap([]*age.Stanza{sshStanza})
		if err != nil {
			plugin.Log.Printf("failed to unwrap file key: %v", err)
			errors = append(errors, plugin.NewInternalErrorStanza(err))
		}
		s := &age.Stanza{
			Type: "file-key",
			Args: []string{strconv.Itoa(i)},
			Body: unwrappedKey,
		}
		stanzas = append(stanzas, s)
	}

	err := respondWithStanzas(stdout, errors, stanzas)
	if err != nil {
		return err
	}

	_, _ = io.WriteString(stdout, "-> done\n\n")
	return nil
}

func RunPlugin(cmd *cobra.Command, _ []string) error {
	switch pluginOptions.AgePlugin {
	case "recipient-v1":
		plugin.Log.Println("Got recipient-v1")
		return RunRecipientV1(os.Stdin, os.Stdout)
	case "identity-v1":
		plugin.Log.Println("Got identity-v1")
		return RunIdentityV1(os.Stdin, os.Stdout)
	default:
		return RunCli(cmd, os.Stdout)
	}
}

func pluginFlags(cmd *cobra.Command, _ *PluginOptions) error {
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVarP(&pluginOptions.OutputFile, "output", "o", "", "Write the result to the file.")

	flags.StringVarP(&pluginOptions.Generate, "generate", "g", "", "Generate a identity based on a 1Password SSH key.")

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
