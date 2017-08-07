#!/bin/sh

# {{.name}}: {{.description}}
#
# {{.copyright}}
#
# {{.license}}
#

aclocal -I m4 &&
	libtoolize --automake && \
	autoheader && \
	automake --foreign --add-missing && \
	autoconf
