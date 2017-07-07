# {{.name}}: {{.description}}
#
# {{.copyright}}
#
# {{.license}}
#
{{if .src_makefile_am}}
{{.src_makefile_am}}{{end}}
if DEBUG
bin_PROGRAMS = {{.name}}d
else
bin_PROGRAMS = {{.name}}
endif

sources ={{range $v := .sources}} \
	{{$v}}{{end}}

{{VarName .name}}d_SOURCES = $(sources)
{{VarName .name}}_SOURCES = $(sources)

{{if .src_extra_dist}}EXTRA_DIST = \
{{.src_extra_dist}}
{{end}}MAINTAINERCLEANFILES = Makefile.in
