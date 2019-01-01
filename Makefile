genDist:
	gox -output "../dist/go_tree_{{.OS}}_{{.Arch}}"

release: genDist
	ghr $(TAG) dist
