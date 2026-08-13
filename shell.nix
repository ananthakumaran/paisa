{ pkgs ? import <nixpkgs> { }, hledger ? import <nixpkgs> { }
, deno ? pkgs.deno, playwright ? pkgs.playwright-driver }:

pkgs.mkShell {
  nativeBuildInputs = [
    pkgs.go_1_24
    pkgs.gotools
    pkgs.gopls
    pkgs.sqlite
    pkgs.nodejs_22
    pkgs.libuuid
    deno
    playwright
    pkgs.node2nix
    # pkgs.pkgsCross.mingwW64.buildPackages.gcc

    pkgs.python312Packages.mkdocs-material
    pkgs.python312Packages.beancount_2

    # test
    pkgs.ledger
    hledger.hledger
  ] ++ (pkgs.lib.optional pkgs.stdenv.isLinux pkgs.wails);

  shellHook = ''
    export CGO_ENABLED=1
    export PLAYWRIGHT_BROWSERS_PATH=${playwright.browsers}
  '';

  env = { LD_LIBRARY_PATH = pkgs.lib.makeLibraryPath [ pkgs.libuuid ]; };
}
