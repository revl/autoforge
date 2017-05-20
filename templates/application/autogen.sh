#!/bin/sh

# {{.PackageName}}: {{.PackageDescription}}
# {{.Copyright}}
#
# {{.License}}
#

aclocal -I m4 &&
	libtoolize --automake && \
	autoheader && \
	automake --foreign --add-missing && \
	autoconf
