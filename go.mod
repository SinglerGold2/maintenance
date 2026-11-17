module github.com/singlergold2/maintenance

go 1.26.11

require github.com/prusa3d/prusa-connect/sdk v0.0.0

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/lib/pq v1.12.3 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/sys v0.48.0 // indirect
	modernc.org/libc v1.77.1 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
	modernc.org/sqlite v1.60.1 // indirect
)

// Until the SDK is published as a Go module, it is built from the pinned
// Prusa Connect checkout that `make core` puts in .core/ (see CORE_REF in
// the Makefile). Delete these lines once `go get` can fetch it.
replace github.com/prusa3d/prusa-connect/sdk => ./.core/sdk
