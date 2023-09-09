class Ledger < Formula
  desc "Command-line, double-entry accounting tool"
  homepage "https://ledger-cli.org/"
  url "https://github.com/ledger/ledger/archive/v3.3.2.tar.gz"
  sha256 "555296ee1e870ff04e2356676977dcf55ebab5ad79126667bc56464cb1142035"
  license "BSD-3-Clause"
  revision 1
  head "https://github.com/ledger/ledger.git", branch: "master"

  livecheck do
    url :stable
    regex(/^v?(\d+(?:\.\d+)+)$/i)
  end

  depends_on "boost"
  depends_on "gmp"
  depends_on "mpfr"
  depends_on "python@3.11" => :build
  depends_on "cmake" => :build

  patch :DATA

  def install
    ENV.cxx11
    ENV.prepend_path "PATH", Formula["python@3.11"].opt_libexec/"bin"

    args = %W[
      --jobs=#{ENV.make_jobs}
      --output=build
      --prefix=#{prefix}
      --boost=#{Formula["boost"].opt_prefix}
      --
      -DCMAKE_FIND_LIBRARY_SUFFIXES=.a
      -DBUILD_SHARED_LIBS:BOOL=OFF
      -DBUILD_DOCS:BOOL=OFF
      -DBoost_NO_BOOST_CMAKE=ON
      -DUSE_PYTHON:BOOL=OFF
      -DUSE_GPGME:BOOL=OFF
      -DBUILD_LIBRARY:BOOL=OFF
      -DBoost_USE_STATIC_LIBS:BOOL=ON
    ] + std_cmake_args
    system "./acprep", "make", *args
    system "./acprep", "make", "install", *args

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
