{
  description = "paisa";
  # Pin hldeger to 1.32.x; 1.40 has a bug related to chained price calculation
  # https://github.com/simonmichael/hledger/issues/2254
  inputs.hledger-pkgs.url =
    "github:NixOS/nixpkgs/ebe4301cbd8f81c4f8d3244b3632338bbeb6d49c";

  outputs = { self, nixpkgs, flake-utils, hledger-pkgs }:
    flake-utils.lib.eachSystem [
      "x86_64-linux"
      "aarch64-linux"
      "aarch64-darwin"
    ] (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = "0.7.5";
        hledger = hledger-pkgs.legacyPackages.${system};
        frontend = pkgs.buildNpmPackage {
          pname = "paisa-frontend";
          inherit version;
          src = ./.;

          nativeBuildInputs = [ pkgs.nodejs_24 ];
          npmDepsHash = "sha256-kG8oqDf7ZkbcXGwbNNMNjzIQMU4ec3lSQeXYxwkh1wo=";
          npmBuildScript = "build";
          npmInstallFlags = [ "--ignore-scripts" ];
          npmRebuildFlags = [ "--ignore-scripts" ];

          installPhase = ''
            mkdir -p $out
            cp -r web/static $out/
          '';
        };
      in {
        packages.frontend = frontend;
        devShells.default = import ./shell.nix {
          inherit pkgs;
          inherit hledger;
        };

        packages.default = pkgs.buildGoModule {
          pname = "paisa-cli";
          meta.mainProgram = "paisa";
          inherit version;

          src = ./.;

          nativeBuildInputs = [ pkgs.nodejs_24 ];

          vendorHash = "sha256-5jrxI+zSKbopGs5GmGVyqQcMHNZJbCsiFEH/LPXWxpk=";

          env = {
            CGO_ENABLED = 1;
          };

          doCheck = false;

          subPackages = [ "." ];

          preConfigure = ''
            cp -r ${frontend}/static web/static
          '';

        };
      });
}
