{ pkgs ? import <nixos-unstable> { } }:

let
  mdbook-admonish = pkgs.rustPlatform.buildRustPackage rec {
    pname = "mdbook-admonish";
    version = "1.4.0";

    src = pkgs.fetchFromGitHub {
      owner = "tommilligan";
      repo = "mdbook-admonish";
      rev = "v${version}";
      sha256 = "sha256:1kiyz78zmiw066bf7v9az0zq1nmk5aqwlj5bd2aij3r7c353lx9a";
    };

    cargoSha256 = "sha256:0yxkvbxgxzhchnx114icvswkag355rf5nhni8xif8hs1x7y3b13h";
  };
in pkgs.mkShell {
  nativeBuildInputs = [
    pkgs.go_1_18
    pkgs.gopls
    pkgs.sqlite
    pkgs.nodejs-17_x
    pkgs.mdbook
    mdbook-admonish
  ];

  shellHook = ''
    export CGO_ENABLED=1
  '';
}
