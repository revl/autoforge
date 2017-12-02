// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

var appTemplate = []embeddedTemplateFile{
	embeddedTemplateFile{"configure.ac", 0644,
		[]byte(`{{template "FileHeader" . -}}
AC_INIT([{{.name}}], [{{.version}}])
AC_CONFIG_AUX_DIR([config])
AC_CONFIG_MACRO_DIRS([m4])
{{$sources := Dir "src" -}}
{{if eq (len $sources) 0}}
{{Error "The app template requires at least one source file in src/"}}
{{end -}}
AC_CONFIG_SRCDIR([src/{{index $sources 0}}])
AC_CONFIG_HEADERS([config.h])
AM_INIT_AUTOMAKE([foreign])

test -z "$CXXFLAGS" && CXXFLAGS=""

AC_PROG_CXX
LT_INIT([disable-shared])

dnl When compiling with GNU C++, display more warnings.
AS_IF([test "$GXX" = yes],
	[CXXFLAGS="$CXXFLAGS -ansi -pedantic -Wall \
-Woverloaded-virtual -Wsign-promo -W -Wshadow -Wpointer-arith -Wcast-qual \
-Wwrite-strings -Wconversion -Wsign-compare -Wredundant-decls -Winline"],
dnl Display all levels of the Digital (Compaq) C++ warnings.
[test "$CXX" = cxx &&
	cxx -V < /dev/null 2>&1 | grep -Eiq 'digital|compaq'],
	[DIGITALCXX="yes"
	CXXFLAGS="$CXXFLAGS -w0 -msg_display_tag -std strict_ansi"],
dnl Enable all warnings and remarks of the Intel C++ compiler.
[test "$CXX" = icpc && icpc -V < /dev/null 2>&1 | grep -iq intel],
	[CXXFLAGS="$CXXFLAGS -w2"])

AC_ARG_ENABLE(debug, AS_HELP_STRING([--enable-debug],
	[enable debug info and runtime checks (default=no)]))

AM_CONDITIONAL(DEBUG, [test "$enable_debug" = yes])

AS_IF([test "$enable_debug" != yes],
	[CXXFLAGS="$CXXFLAGS -O3"],
[AC_DEFINE([DEBUG], 1, [Define to 1 to enable various runtime checks.])
AS_IF([test "$GXX" = yes],
	[CXXFLAGS="$CXXFLAGS -ggdb"],
[test "$DIGITALCXX" = yes],
	[CXXFLAGS="$CXXFLAGS -gall"],
[test "$ac_cv_prog_cxx_g" = yes],
	[CXXFLAGS="$CXXFLAGS -g"])])
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
{{end}}{{end -}}
{{template "Snippet" .}}
AC_CONFIG_FILES([Makefile
src/Makefile])
AC_OUTPUT
`)},
	embeddedTemplateFile{"Makefile.am", 0644,
		[]byte(`{{template "FileHeader" . -}}
ACLOCAL_AMFLAGS = -I m4

AUTOMAKE_OPTIONS = foreign

SUBDIRS = . src

maintainer-clean-local:
	rm -rf autom4te.cache

EXTRA_DIST = autogen.sh
`)},
	embeddedTemplateFile{"src/Makefile.am", 0644,
		[]byte(`{{template "FileHeader" . -}}
bin_PROGRAMS = {{.name}}

{{$sourceExt := StringList "*?.C" "*?.c" "*?.cc" "*?.cxx" "*?.cpp" -}}
{{$allFiles := Dir .dirname -}}
{{VarName .name -}}
_SOURCES ={{template "Multiline" Select $allFiles $sourceExt}}
{{$extraFiles := Exclude $allFiles $sourceExt -}}
{{if $extraFiles}}
EXTRA_DIST ={{template "Multiline" $extraFiles}}
{{end -}}
{{template "Snippet" .}}`)},
}
