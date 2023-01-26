for os in linux darwin windows; do
	for arch in amd64 arm64;
		do make build_no_tests GOOS="${os}" GOARCH="${arch}"
	done
done
