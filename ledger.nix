{ pkgs ? import <nixpkgs> { } }:

with pkgs;

pkgsStatic.stdenv.mkDerivation {
  pname = "ledger";
  version = "3.3.2";

  src = fetchFromGitHub {
    owner = "ledger";
    repo = "ledger";
    rev = "4355c4faf157d5ef47b126286aa501742732708d";
    hash = "sha256-9IowdrQpJarALr21Y+Mhmld+eC4YUpEn+goqWyBb6Xc=";
  };

  outputs = [ "out" "dev" ];

  buildInputs = [ pkgsStatic.gmp pkgsStatic.mpfr gnused pkgsStatic.boost ];

  nativeBuildInputs = [ cmake tzdata ];

  cmakeFlags = [
    "-DCMAKE_INSTALL_LIBDIR=lib"
    "-DBUILD_DOCS:BOOL=OFF"
    "-DUSE_PYTHON:BOOL=OFF"
    "-DUSE_GPGME:BOOL=OFF"
    "-DBUILD_LIBRARY:BOOL=OFF"
  ];

  enableParallelBuilding = true;

  installTargets = [ "install" ];

  checkPhase = ''
    runHook preCheck
    env LD_LIBRARY_PATH=$PWD \
      DYLD_LIBRARY_PATH=$PWD \
      ctest -j$NIX_BUILD_CORES
    runHook postCheck
  '';

  doCheck = true;

}
