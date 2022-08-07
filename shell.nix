{ pkgs ? import <nixpkgs> { } }:

let
  pinned = import (builtins.fetchTarball {
    url =
      "https://github.com/NixOS/nixpkgs/archive/1ffba9f2f683063c2b14c9f4d12c55ad5f4ed887.tar.gz";
  }) { };

in pkgs.mkShell {
  nativeBuildInputs = [
    pinned.go_1_18
    pinned.gotools
    pinned.gopls
    pkgs.sqlite
    pinned.nodejs-17_x
    pkgs.mdbook
    pkgs.pkgsCross.mingwW64.buildPackages.gcc
  ];

  shellHook = ''
    export CGO_ENABLED=1
  '';
}
