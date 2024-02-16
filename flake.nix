{
  description = "An age plugin using 1Password for encryption keys";

  inputs = { nixpkgs.url = "nixpkgs/nixpkgs-unstable"; };

  outputs = { self, nixpkgs }:
    let
      # To work with older version of flakes
      lastModifiedDate =
        self.lastModifiedDate or self.lastModified or "19700101";

      version = builtins.substring 0 8 lastModifiedDate;

      supportedSystems =
        [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in {
      packages = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          age-plugin-op = pkgs.buildGoModule {
            pname = "age-plugin-op";
            inherit version;
            src = ./.;
            vendorSha256 =
              "sha256-eKeUhS2puz6ALb+cQKl7+DGvm9Cl+miZAHX0imf9wdg=";
          };
        });

      devShells = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [ go gopls gotools go-tools age ];
          };
        });

      defaultPackage =
        forAllSystems (system: self.packages.${system}.age-plugin-op);
    };
}
