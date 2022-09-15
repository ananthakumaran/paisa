{ pkgs ? import <nixpkgs> { } }:

pkgs.mkShell {
  nativeBuildInputs = [
    pkgs.go_1_19
    pkgs.gotools
    pkgs.gopls
    pkgs.sqlite
    pkgs.nodejs-18_x
    pkgs.mdbook
    # pkgs.pkgsCross.mingwW64.buildPackages.gcc
  ];

  shellHook = ''
    export CGO_ENABLED=1
  '';
}
