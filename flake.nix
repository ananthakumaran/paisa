{
  description = "paisa";
  inputs.mkdocs-pkgs.url = "github:NixOS/nixpkgs/staging-next";

  outputs = { self, nixpkgs, flake-utils, mkdocs-pkgs }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        mkdocs = mkdocs-pkgs.legacyPackages.${system};
        nodeDependencies = (pkgs.callPackage ./flake/override.nix {
          nodejs = pkgs.nodejs_22;
        }).nodeDependencies;
      in {
        devShells.default = import ./shell.nix {
          inherit pkgs;
          inherit mkdocs;
        };

        packages.default = pkgs.buildGoModule {
          pname = "paisa-cli";
          meta.mainProgram = "paisa";
          version = "0.7.3";

          src = ./.;

          nativeBuildInputs = [ pkgs.nodejs_22 ];

          vendorHash = "sha256-KnHJ6+aMahTeNdbRcRAgBERGVYen/tM/tDcFI/NyLdE=";

          CGO_ENABLED = 1;

          doCheck = false;

          subPackages = [ "." ];

          preConfigure = ''
            ln -s ${nodeDependencies}/lib/node_modules ./node_modules
            export PATH="${nodeDependencies}/.bin:$PATH"
            npm run build
          '';

        };
      });
}
