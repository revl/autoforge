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
package definition file, whose name must consists of the name of the
package directory and the extension `.yaml`.

Among other parameters, the package definition file specifies a template
that the package uses, which, in turn, determines the type of binary
that the package produces.

Because project templates encapsulate a great deal of complexity that
comes with using Autotools, the structure of the package definition file
is quite simple, which makes starting a new project a breeze.

## Project templates

Project templates contain autoconf and automake source files required
for building the project. Autoforge provides several generic templates.
Additional templates can be created ad hoc.

## Project definition files

By imposing certain restrictions on the project structure, Autoforge
keeps the differences between the projects generated from the same
template to a minimum. For a new project, its package definition file
must specify just a few crucial parameters, such as the name of the
project, the type of the license it uses, etc. A separate section
below describes the full list of parameters that can appear in a
package definition file.

## Package search path

The `AUTOFORGE_PKG_PATH` environment variable defines a colon-separated
list of directories that contain packages.  Autoforge searches for
package definition files in subdirectories of the `AUTOFORGE_PKG_PATH`
directories. Subdirectories without such files are ignored.

## The list of package definition file parameters

Here is the full list of variables that can appear in a package
definition file:

- `name`

  The name of the package. This name does not have to match the name
  of the directory that contains the package.

- `template`

  The name of the package template. At the moment, either `library` or
  `application`.

- `version`

  Package version for use by Automake.

- `version_info`

  API/ABI revision for use by Libtool.

- `requires`

  The list of libraries that the package requires.

- `headers`

  For a library, the list of C/C++ headers exported by the library.

- `sources`

  The list of C/C++ sources containing the implementation.

- `configure`

  A snippet to be embedded in the `configure.in` file. Can be a mix of
  Bourne shell code and Autoconf macros.
