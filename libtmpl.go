// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

var libTemplate = []embeddedTemplateFile{
	{"include/{name}/Makefile.am", 0644,
		[]byte(`{{template "FileHeader" . -}}
pkgincludedir = $(includedir)/{{.name}}

{{$headerExt := StringList "*?.H" "*?.h" "*?.hh" "*?.hxx" "*?.hpp" -}}
{{$allFiles := Dir .dirname -}}
pkginclude_HEADERS ={{template "Multiline" Select $allFiles $headerExt}}
{{$extraFiles := Exclude $allFiles $headerExt}}{{if $extraFiles}}
EXTRA_DIST ={{template "Multiline" $extraFiles}}
{{end}}
all: config.h

check: all

config.h: $(CONFIG_HEADER)
	sed -e 's/#\([^ ][^ ]*\) \([^ ][^ ]*\)/#\1 {{VarNameUC .name}}_\2/g' < \
		$(CONFIG_HEADER) > $@

install-data-hook: config.h
	$(INSTALL_DATA) config.h "$(DESTDIR)$(pkgincludedir)/config.h"

uninstall-hook:
	rm -f "$(DESTDIR)$(pkgincludedir)/config.h"
	rmdir "$(DESTDIR)$(pkgincludedir)" || true

# For older versions of Automake.
uninstall-local:
	rm -f "$(DESTDIR)$(pkgincludedir)/config.h"
	rmdir "$(DESTDIR)$(pkgincludedir)" || true

CLEANFILES = config.h
{{template "Snippet" .}}`)},
	{"include/Makefile.am", 0644,
		[]byte(`{{template "FileHeader" . -}}
SUBDIRS = {{.name}}
{{template "Snippet" .}}`)},
	{"src/Makefile.am", 0644,
		[]byte(`{{template "FileHeader" . -}}
lib_LTLIBRARIES = lib{{.name}}.la

{{if index . "version-info" -}}
lib{{VarName .name}}_la_LDFLAGS = -version-info @library_version_info@

{{end -}}
{{$sourceExt := StringList "*?.C" "*?.c" "*?.cc" "*?.cxx" "*?.cpp" -}}
{{$allFiles := Dir .dirname -}}
lib{{VarName .name -}}
_la_SOURCES ={{template "Multiline" Select $allFiles $sourceExt}}
{{$extraFiles := Exclude $allFiles $sourceExt -}}
{{if $extraFiles}}
EXTRA_DIST ={{template "Multiline" $extraFiles}}
{{end -}}
{{template "Snippet" .}}`)},
	{"tests/Makefile.am", 0644,
		[]byte(`{{template "FileHeader" . -}}
LDADD = ../src/lib$(PACKAGE).la

{{$sourceExt := StringList "*?.C" "*?.c" "*?.cc" "*?.cxx" "*?.cpp" -}}
{{$allFiles := Dir .dirname -}}
{{$testSources := Select $allFiles $sourceExt -}}
{{if eq (len $testSources) 0}}
{{Error "'lib' template requires at least one test_*.{cc,c} file under tests/"}}
{{end -}}
check_PROGRAMS ={{range $testSources}} \
	{{TrimExt .}}{{end}}

{{range $testSources -}}
{{VarName (TrimExt .)}}_SOURCES = {{.}}

{{end -}}
TESTS = $(check_PROGRAMS)
{{$extraFiles := Exclude $allFiles $sourceExt -}}
{{if $extraFiles}}
EXTRA_DIST ={{template "Multiline" $extraFiles}}
{{end -}}
{{template "Snippet" .}}`)},
	{"Makefile.am", 0644,
		[]byte(`{{template "FileHeader" . -}}
{{if gt (len (Dir "m4")) 0 -}}
ACLOCAL_AMFLAGS = -I m4

{{end -}}
AUTOMAKE_OPTIONS = foreign

SUBDIRS = . include src tests

pkgconfig_DATA = {{.name}}.pc

EXTRA_DIST = autogen.sh
{{template "Snippet" .}}`)},
	{"configure.ac", 0644,
		[]byte(`{{template "FileHeader" . -}}
AC_INIT([{{.name}}], [{{.version}}])
AC_CONFIG_AUX_DIR([config])
{{if gt (len (Dir "m4")) 0 -}}
AC_CONFIG_MACRO_DIRS([m4])
{{end -}}
{{$sources := Dir "src" -}}
{{if eq (len $sources) 0}}
{{Error "'lib' template requires at least one source file in src/"}}
{{end -}}
AC_CONFIG_SRCDIR([src/{{index $sources 0}}])
AC_CONFIG_HEADERS([config.h])
AM_INIT_AUTOMAKE([foreign])

{{if index . "version-info" -}}
library_version_info={{index . "version-info"}}
AC_SUBST(library_version_info)

{{end -}}
test -z "$CXXFLAGS" && CXXFLAGS=""

AC_PROG_CXX
LT_INIT([disable-shared])
PKG_PROG_PKG_CONFIG
PKG_INSTALLDIR

CPPFLAGS="$CPPFLAGS -I\$(top_srcdir)/include -I\$(top_builddir)/include"

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
[CPPFLAGS="$CPPFLAGS -D{{VarNameUC .name}}_DEBUG"
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
AC_SUBST(CONFIG_FLAGS)
AC_SUBST(CONFIG_LIBS)
AC_SUBST(PRIVATE_CONFIG_LIBS)
AC_SUBST(UNINST_PREFIX)
AC_SUBST(UNINST_FLAGS)
AC_SUBST(UNINST_LIBS)

AC_CONFIG_FILES([Makefile
include/Makefile
include/{{.name}}/Makefile
src/Makefile
tests/Makefile
{{.name}}.pc
{{.name}}-uninstalled.pc])
AC_OUTPUT
`)},
	{"{name}-uninstalled.pc.in", 0644,
		[]byte(`prefix=@UNINST_PREFIX@
exec_prefix=@UNINST_PREFIX@
libdir=@UNINST_PREFIX@/src
includedir=@UNINST_PREFIX@/include

Name: @PACKAGE_NAME@
Description: {{.description}}
Version: @PACKAGE_VERSION@
Libs: @UNINST_LIBS@
Libs.private: @PRIVATE_CONFIG_LIBS@
Cflags: @UNINST_FLAGS@
`)},
	{"{name}.pc.in", 0644,
		[]byte(`prefix=@prefix@
exec_prefix=@exec_prefix@
libdir=@libdir@
includedir=@includedir@

Name: @PACKAGE_NAME@
Description: {{.description}}
Version: @PACKAGE_VERSION@
Libs: @CONFIG_LIBS@
Libs.private: @PRIVATE_CONFIG_LIBS@
Cflags: @CONFIG_FLAGS@
`)},
}
