{
  description = "paisa";
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        nodeDependencies = (pkgs.callPackage ./flake/override.nix {
          nodejs = pkgs.nodejs-18_x;
        }).nodeDependencies;
      in {
        devShells.default = import ./shell.nix { inherit pkgs; };

        packages.default = pkgs.buildGoModule {
          pname = "paisa-cli";
          version = "0.6.2";

          src = ./.;

          nativeBuildInputs = [ pkgs.nodejs-18_x ];

          vendorHash = "sha256-9VM6+0h6DMtlA+RvCfpDffAvaq7jRmBVVFTbdUfYwsA=";

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
