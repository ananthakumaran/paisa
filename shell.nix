{ pkgs ? import <nixpkgs> { } }:

pkgs.mkShell {
  nativeBuildInputs = [
    pkgs.go_1_20
    pkgs.gotools
    pkgs.gopls
    pkgs.sqlite
    pkgs.nodejs-18_x
    pkgs.libuuid
    # pkgs.pkgsCross.mingwW64.buildPackages.gcc

    pkgs.python311Packages.mkdocs
    pkgs.python311Packages.mkdocs-material
  ] ++ (pkgs.lib.optional pkgs.stdenv.isLinux pkgs.wails);

  shellHook = ''
    export CGO_ENABLED=1
  '';

  env = { LD_LIBRARY_PATH = pkgs.lib.makeLibraryPath [ pkgs.libuuid ]; };
}
