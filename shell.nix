{ pkgs ? import <nixpkgs> { }, mkdocs ? import <nixpkgs> { } }:

pkgs.mkShell {
  nativeBuildInputs = [
    pkgs.go_1_21
    pkgs.gotools
    pkgs.gopls
    pkgs.sqlite
    pkgs.nodejs-18_x
    pkgs.libuuid
    pkgs.bun
    pkgs.node2nix
    # pkgs.pkgsCross.mingwW64.buildPackages.gcc

    mkdocs.python311Packages.mkdocs-material
    pkgs.python311Packages.beancount

    # test
    pkgs.ledger
    pkgs.hledger
  ] ++ (pkgs.lib.optional pkgs.stdenv.isLinux pkgs.wails);

  shellHook = ''
    export CGO_ENABLED=1
  '';

  env = { LD_LIBRARY_PATH = pkgs.lib.makeLibraryPath [ pkgs.libuuid ]; };
}
