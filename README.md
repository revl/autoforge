----

_This project is under active development, but it's far from being ready.
Please do not attempt to make use of it until this notice disappears._

----

# autoforge

This utility is an attempt to implement a higher level build system on
top of GNU Autotools. Autoforge generates autoconf scripts and automake
source files from whole project templates. It also tracks inter-project
dependencies and creates a meta-Makefile to build the projects in the
correct order.

## Packages

An Autoforge package is a directory containing C/C++ source files and a
project definition file. The name of the file must match the name of the
directory, and the file extension must be `.yaml`. Among other
parameters, the project definition file specifies a template that the
package uses, which, in turn, determines the type of binary that the
package produces.

Because project templates encapsulate a great deal of complexity that
comes with using Autotools, the structure of the project definition
file is quite simple, which makes starting a new project a breeze.

## Project templates

Project templates contain autoconf and automake source files required
for building the project. Autoforge provides several generic templates.
Additional templates can be created ad hoc.

## Project definition files

By imposing certain restrictions on the project structure, Autoforge
keeps the differences between the projects generated from the same
template to a minimum. For a new project, its package definition file
must specify just a few crucial parameters, such as the name of the
project, the type of the license it uses, etc.

## Package search path

The `AUTOFORGE_PKG_PATH` environment variable defines a colon-separated
list of directories that contain packages.  Autoforge searches for
package definition files in subdirectories of the `AUTOFORGE_PKG_PATH`
directories. Subdirectories without such files are ignored.
