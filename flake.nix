{
  description = "paisa";
  # Pin hldeger to 1.32.x; 1.40 has a bug related to chained price calculation
  # https://github.com/simonmichael/hledger/issues/2254
  inputs.hledger-pkgs.url =
    "github:NixOS/nixpkgs/ebe4301cbd8f81c4f8d3244b3632338bbeb6d49c";

  outputs = { self, nixpkgs, flake-utils, hledger-pkgs }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        hledger = hledger-pkgs.legacyPackages.${system};
        nodeDependencies = (pkgs.callPackage ./flake/override.nix {
          nodejs = pkgs.nodejs_22;
        }).nodeDependencies;
      in {
        devShells.default = import ./shell.nix {
          inherit pkgs;
          inherit hledger;
        };

        packages.default = pkgs.buildGoModule {
          pname = "paisa-cli";
          meta.mainProgram = "paisa";
          version = "0.7.3";

          src = ./.;

          nativeBuildInputs = [ pkgs.nodejs_22 ];

          vendorHash = "sha256-5jrxI+zSKbopGs5GmGVyqQcMHNZJbCsiFEH/LPXWxpk=";

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
