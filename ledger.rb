class Ledger < Formula
  desc "Command-line, double-entry accounting tool"
  homepage "https://ledger-cli.org/"
  license "BSD-3-Clause"
  revision 6
  head "https://github.com/ledger/ledger.git", branch: "master"

  stable do
    url "https://github.com/ledger/ledger/archive/refs/tags/v3.3.2.tar.gz"
    sha256 "555296ee1e870ff04e2356676977dcf55ebab5ad79126667bc56464cb1142035"

    # Support building with mandoc
    # Remove with v3.4.x
    patch do
      url "https://github.com/ledger/ledger/commit/f40cee6c3af4c9cec05adf520fc7077a45060434.patch?full_index=1"
      sha256 "d5be89dbadff7e564a750c10cdb04b83e875452071a2115dd70aae6e7a8ee76c"
    end
    patch do
      url "https://github.com/ledger/ledger/commit/14b90d8d952b40e0a474223e7f74a1e6505d5450.patch?full_index=1"
      sha256 "d250557e385163e3ad3002117ebe985af040d915aab49ae1ea342db82398aeda"
    end

    # Backport fix to build with `boost` 1.85.0
    patch do
      url "https://github.com/ledger/ledger/commit/46207852174feb5c76c7ab894bc13b4f388bf501.patch?full_index=1"
      sha256 "8aaf8daf4748f359946c64488c96345f4a4bdf928f6ec7a1003610174428599f"
    end
  end

  livecheck do
    url :stable
    regex(/^v?(\d+(?:\.\d+)+)$/i)
  end

  depends_on "boost@1.85"
  depends_on "gmp"
  depends_on "mpfr"
  depends_on "python@3.12" => :build
  depends_on "cmake" => :build

  patch :DATA

  def install
    ENV.cxx11
    ENV.prepend_path "PATH", Formula["python@3.12"].opt_libexec/"bin"

    args = %W[
      --jobs=#{ENV.make_jobs}
      --output=build
      --prefix=#{prefix}
      --boost=#{Formula["boost@1.85"].opt_prefix}
      --
      -DCMAKE_FIND_LIBRARY_SUFFIXES=.a
      -DCMAKE_POLICY_VERSION_MINIMUM=3.5
      -DBUILD_SHARED_LIBS:BOOL=OFF
      -DBUILD_DOCS:BOOL=OFF
      -DBoost_NO_BOOST_CMAKE=ON
      -DUSE_PYTHON:BOOL=OFF
      -DUSE_GPGME:BOOL=OFF
      -DBUILD_LIBRARY:BOOL=OFF
      -DBoost_USE_STATIC_LIBS:BOOL=ON
    ] + std_cmake_args

    system "./acprep", "opt", "make", *args
    system "./acprep", "opt", "make", "install", *args
  end

  test do
    balance = testpath/"output"
    system bin/"ledger",
      "--args-only",
      "--file", pkgshare/"examples/sample.dat",
      "--output", balance,
      "balance", "--collapse", "equity"
    assert_equal "          $-2,500.00  Equity", balance.read.chomp
  end
end

__END__
diff --git a/CMakeLists.txt b/CMakeLists.txt
index 83a6f89d..e84be891 100644
--- a/CMakeLists.txt
+++ b/CMakeLists.txt
@@ -274,8 +274,8 @@ find_opt_library_and_header(EDIT_PATH histedit.h EDIT_LIB edit HAVE_EDIT)
 ########################################################################

 macro(add_ledger_library_dependencies _target)
-  target_link_libraries(${_target} ${MPFR_LIB})
-  target_link_libraries(${_target} ${GMP_LIB})
+  target_link_libraries(${_target} /usr/local/lib/libmpfr.a)
+  target_link_libraries(${_target} /usr/local/lib/libgmp.a)
   if (HAVE_EDIT)
     target_link_libraries(${_target} ${EDIT_LIB})
   endif()
