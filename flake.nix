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
    ] (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = "0.8.1";
        hledger = hledger-pkgs.legacyPackages.${system};
        frontend = pkgs.buildNpmPackage {
          pname = "paisa-frontend";
          inherit version;
          src = ./.;

          nativeBuildInputs = [ pkgs.nodejs_24 pkgs.makeWrapper ];
          npmDepsHash = "sha256-481E+4KVeZwSTeZxCYmxaurR5ndHQzJeqxEYhjZiZLs=";
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
        devShells.beancount3 = import ./shell.nix {
          inherit pkgs hledger;
          beancount3 = true;
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

          postInstall = ''
            wrapProgram $out/bin/paisa \
              --prefix PATH : ${pkgs.lib.makeBinPath [ hledger.hledger ]}
          '';

          preConfigure = ''
            cp -r ${frontend}/static web/static
          '';

        };
      });
}
