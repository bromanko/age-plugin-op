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
          age-plugin-op = pkgs.callPackage ./nix/age-plugin-op.nix { };

          age-plugin-op-linux-arm64 =
            (pkgs.callPackage ./nix/age-plugin-op.nix { }).overrideAttrs (old: {
              GOOS = "linux";
              GOARCH = "arm64";
            });
          age-plugin-op-linux-arm32 =
            (pkgs.callPackage ./nix/age-plugin-op.nix { }).overrideAttrs (old: {
              GOOS = "linux";
              GOARCH = "arm";
              GOARM = "6";
            });
          age-plugin-op-linux-amd64 =
            (pkgs.callPackage ./nix/age-plugin-op.nix { }).overrideAttrs (old: {
              GOOS = "linux";
              GOARCH = "amd64";
            });
          age-plugin-op-darwin-amd64 =
            (pkgs.callPackage ./nix/age-plugin-op.nix { }).overrideAttrs (old: {
              GOOS = "darwin";
              GOARCH = "amd64";
            });
          age-plugin-op-darwin-arm64 =
            (pkgs.callPackage ./nix/age-plugin-op.nix { }).overrideAttrs (old: {
              GOOS = "darwin";
              GOARCH = "arm64";
            });
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
