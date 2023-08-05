{ pkgs ? import <nixpkgs> { } }:

pkgs.mkShell {
  nativeBuildInputs = [
    pkgs.go_1_20
    pkgs.gotools
    pkgs.gopls
    pkgs.sqlite
    pkgs.nodejs-18_x
    pkgs.mdbook
    pkgs.libuuid
    pkgs.electron_25
    pkgs.pkg-config
    pkgs.pixman
    pkgs.cairo
    pkgs.pango
    pkgs._7zz
    pkgs.flatpak-builder
    pkgs.flatpak
    pkgs.elfutils
    # pkgs.pkgsCross.mingwW64.buildPackages.gcc
  ];

  shellHook = ''
    export CGO_ENABLED=1
  '';

  env = { LD_LIBRARY_PATH = pkgs.lib.makeLibraryPath [ pkgs.libuuid ]; };
}
