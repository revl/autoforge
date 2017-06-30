# {{.PackageName}}: {{.PackageDescription}}
#
# {{.Copyright}}
#
# {{.License}}
#
{{if .SRC_MAKEFILE_AM}}
{{.SRC_MAKEFILE_AM}}{{end}}
if DEBUG
bin_PROGRAMS = {{.PackageName}}d
else
bin_PROGRAMS = {{.PackageName}}
endif

sources ={{range $v := .SOURCES}} \
	{{$v}}{{end}}

{{VarName .PackageName}}d_SOURCES = $(sources)
{{VarName .PackageName}}_SOURCES = $(sources)

{{if .SRC_EXTRA_DIST}}EXTRA_DIST = \
{{.SRC_EXTRA_DIST}}
{{end}}MAINTAINERCLEANFILES = Makefile.in
