{ pkgs ? import <nixpkgs> { }, hledger ? import <nixpkgs> { }, beancount3 ? false }:

pkgs.mkShell {
  nativeBuildInputs = [
    pkgs.go_1_26
    pkgs.gotools
    pkgs.gopls
    pkgs.sqlite
    pkgs.nodejs_24
    pkgs.libuuid
    # pkgs.pkgsCross.mingwW64.buildPackages.gcc

    pkgs.python312Packages.mkdocs-material

    # test
    pkgs.ledger
    hledger.hledger
  ] ++ (if beancount3 then [
    pkgs.python312Packages.beancount
    pkgs.python312Packages.beanquery
  ] else [
    pkgs.python312Packages.beancount_2
  ]) ++ (pkgs.lib.optional pkgs.stdenv.hostPlatform.isLinux pkgs.wails);

  shellHook = ''
    export CGO_ENABLED=1
  '';

  env = { LD_LIBRARY_PATH = pkgs.lib.makeLibraryPath [ pkgs.libuuid ]; };
}
