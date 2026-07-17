.PHONY: test copy update remove

test: copy update remove
copy:
	mkdir -p test; cp -r .test/* test/
update:
	go run ./cmd/galaxy-update -v test/requirements.yml
remove:
	rm -rf test
