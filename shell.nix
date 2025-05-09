{ pkgs ? import <nixpkgs> { }, mkdocs ? import <nixpkgs> { } }:

pkgs.mkShell {
  nativeBuildInputs = [
    pkgs.go_1_24
    pkgs.gotools
    pkgs.gopls
    pkgs.sqlite
    pkgs.nodejs_22
    pkgs.libuuid
    pkgs.bun
    pkgs.node2nix
    # pkgs.pkgsCross.mingwW64.buildPackages.gcc

    mkdocs.python312Packages.mkdocs-material
    pkgs.python312Packages.beancount_2

    # test
    pkgs.ledger
    pkgs.hledger
  ] ++ (pkgs.lib.optional pkgs.stdenv.isLinux pkgs.wails);

  shellHook = ''
    export CGO_ENABLED=1
  '';

  env = { LD_LIBRARY_PATH = pkgs.lib.makeLibraryPath [ pkgs.libuuid ]; };
}
