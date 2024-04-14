> *⚠️  I have been using this plugin for a while without issue.
> However, it hasn't received much review. YMMV.*

# 1Password Plugin for Age Clients

`age-plugin-op` is a plugin for age clients like [age](https://github.com/FiloSottile/age/)
and [rage](https://github.com/str4d/rage). It allows you to use your 1Password SSH keys with age clients.

## Requirements

- The [1Password CLI](https://1password.com/downloads/command-line/) installed on `$PATH`
- An age client such as [age](https://github.com/FiloSottile/age/) or [rage](https://github.com/str4d/rage)

## Installation

Age plugins must be available in the `$PATH`. Right now there are no pre-built binaries, so you will need to build the plugin yourself.

## Usage

1. Generate a new age identity file using the `age-plugin-op` plugin. You will be prompted to authenticate with 1Password.
    ```sh
    $ age-plugin-op --generate "op://Personal/wxrzetxonuggniebjzruxycq/private key" -o age-identity.txt
    # Created: 2024-02-16 13:25:00.433868 -0800 PST m=+0.003075709
    # Recipient: age1op102xjaf99y9u69cf64cl8trptuenerd3gal8t4hc2exd8z4ntvpyquwaf9l

    AGE-PLUGIN-OP-1Q9D7XC8RDFW0X3F9P7R9WGZDTST5V22CQUMUM3MK6VTKWKJ[...]
   ```
2. Encrypt a file/stream for the recipient.
    ```sh
    $ echo "Hello World" | age -r "age1op102xjaf99y9u69cf64cl8trptuenerd3gal8t4hc2exd8z4ntvpyquwaf9l" > secret.age
   ```
3. Decrypt the file/stream using the `age-plugin-op` plugin.
    ```sh
    $ age --decrypt -i age-identity.txt -o - secret.age
    Hello World`
    ```

## Supported SSH Key Types

Only RSA and ECD25519 keys are supported since these are the only types [supported by 1Password](https://developer.1password.com/docs/ssh/agent/#eligible-keys).

## Inspiration

This plugin is heavily inspired by the [age-plugin-tpm](https://github.com/Foxboron/age-plugin-tpm/) plugin and [age-plugin-se](https://github.com/remko/age-plugin-se/) plugin.
Some internal code from age was used directly.

## License

Licensed under the MIT license. See [LICENSE](LICENSE) or http://opensource.org/licenses/MIT

