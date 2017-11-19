// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

var appTemplate = []embeddedTemplateFile{
	embeddedTemplateFile{"configure.ac", 0644,
		[]byte(`{{template "FileHeader" . -}}
AC_INIT([{{.name}}], [{{.version}}])
AC_CONFIG_AUX_DIR([config])
AC_CONFIG_MACRO_DIRS([m4]){{if .sources}}
AC_CONFIG_SRCDIR([src/{{index .sources 0}}]){{else}}
{{$ss := Dir "src"}}{{if eq (len $ss) 0}}
	{{Error "The app template requires at least one source file in src/"}}
{{end}}AC_CONFIG_SRCDIR([src/{{index $ss 0}}]){{end}}
AC_CONFIG_HEADERS([config.h])
AM_INIT_AUTOMAKE([foreign])

test -z "$CXXFLAGS" && CXXFLAGS=""

dnl Checks for programs.
AC_PROG_CXX
AC_PROG_LIBTOOL

if test "x$GXX" = "xyes"; then
	CXXFLAGS="$CXXFLAGS -Wall"
elif test "$CXX" = cxx && cxx -V < /dev/null 2>&1 | \
	grep -Eiq 'digital|compaq'; then
	DIGITALCXX="yes"
	CXXFLAGS="$CXXFLAGS -w0 -msg_display_tag -std ansi -nousing_std"
	CXXFLAGS="$CXXFLAGS -D__USE_STD_IOSTREAM -D_POSIX_PII_SOCKET"
fi
{{if .snippets}}{{if index .snippets "configure.ac"}}
{{index .snippets "configure.ac"}}{{end}}{{end}}
ACX_PTHREAD(,[AC_MSG_ERROR([this package requires pthreads support])])

CXXFLAGS="$CXXFLAGS $PTHREAD_CFLAGS"
LIBS="$LIBS $PTHREAD_LIBS"

AC_ARG_ENABLE(debug, changequote(<<, >>)<<  --enable-debug          >>dnl
<<enable debug info and runtime checks [default=no]>>changequote([, ]))

AM_CONDITIONAL(DEBUG, test "$enable_debug" = yes)

if test "$enable_debug" != yes; then
	CXXFLAGS="$CXXFLAGS -O3"
elif test "$DIGITALCXX" = yes; then
	CXXFLAGS="$CXXFLAGS -D_DEBUG -gall"
elif test "$GXX" = yes; then
	CXXFLAGS="$CXXFLAGS -D_DEBUG -ggdb"
elif test "$ac_cv_prog_cxx_g" = yes; then
	CXXFLAGS="$CXXFLAGS -D_DEBUG -g"
fi
{{if or .external_libs .requires}}
dnl Checks for libraries.{{end}}{{if .external_libs}}{{range .external_libs}}
AC_CHECK_LIB([{{.name}}], [{{.function}}],,
	AC_MSG_ERROR([unable to link with {{.name}}]){{if .other_libs}},
	[{{.other_libs}}]{{end}}){{end}}
{{end}}{{if .requires}}
PKG_PROG_PKG_CONFIG()
{{range .requires}}
PKG_CHECK_MODULES([{{VarNameUC .}}], [{{VarName .}}])
CXXFLAGS="$CXXFLAGS ${{VarNameUC .}}_CFLAGS"
LIBS="$LIBS ${{VarNameUC .}}_LIBS"
{{end}}{{end}}
AC_OUTPUT([Makefile
src/Makefile
config/Makefile
m4/Makefile])
`)},
	embeddedTemplateFile{"config/Makefile.am", 0644,
		[]byte(`{{template "FileHeader" . -}}
MAINTAINERCLEANFILES = \
	config.guess \
	config.sub \
	depcomp \
	install-sh \
	ltconfig \
	mkinstalldirs \
	test-driver \
	Makefile.in
`)},
	embeddedTemplateFile{"Makefile.am", 0644,
		[]byte(`{{template "FileHeader" . -}}
ACLOCAL_AMFLAGS = -I m4

AUTOMAKE_OPTIONS = foreign

SUBDIRS = . config m4 src

maintainer-clean-local:
	rm -rf autom4te.cache

EXTRA_DIST = autogen.sh

MAINTAINERCLEANFILES = Makefile.in
`)},
	embeddedTemplateFile{"src/Makefile.am", 0644,
		[]byte(`{{template "FileHeader" . -}}
{{if .snippets}}{{if index .snippets "src/Makefile.am" -}}
{{index .snippets "src/Makefile.am"}}
{{end}}{{end}}if DEBUG
bin_PROGRAMS = {{.name}}d
else
bin_PROGRAMS = {{.name}}
endif{{$srcFileTypes := StringList "*?.C" "*?.c" "*?.cc" "*?.cxx" "*?.cpp"}}

sources ={{if .sources}}{{template "Multiline" .sources}}
{{else}}{{template "Multiline" Select (Dir "src") $srcFileTypes}}
{{end}}
{{VarName .name}}d_SOURCES = $(sources)
{{VarName .name}}_SOURCES = $(sources)
{{if .src_extra_dist}}
EXTRA_DIST ={{template "Multiline" .src_extra_dist}}
{{else}}{{$extraFiles := Exclude (Dir "src") $srcFileTypes}}{{if $extraFiles}}
EXTRA_DIST ={{template "Multiline" $extraFiles}}
{{end}}{{end}}
MAINTAINERCLEANFILES = Makefile.in
`)},
}
