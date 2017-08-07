# {{.name}}: {{.description}}
#
# {{.copyright}}
#
# {{.license}}
#

ACLOCAL_AMFLAGS = -I m4

AUTOMAKE_OPTIONS = foreign

SUBDIRS = . config m4 src

maintainer-clean-local:
	rm -rf autom4te.cache

EXTRA_DIST = autogen.sh

MAINTAINERCLEANFILES = Makefile.in
