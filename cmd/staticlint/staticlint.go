// This is a static checker for go
//
// # Usage
//
// Build and add to $PATH
//
//	go vet -vettool=staticlint [options] ./...
//
// # Options
//
//	staticlint help
//
// # Checkers
//
// The following checkers are included:
//
//   - exitcheck: no direct os.Exit calls allowed in func main() of package main
//   - ineffassign: if the variable assigned is not thereafter used
//   - bodyclose: bodyclose checks whether HTTP response body is closed successfully
//   - all standard checks of go vet tool: https://pkg.go.dev/cmd/vet
//   - all SA staticcheck: https://staticcheck.dev/docs/checks/#SA
//   - QF1002: staticcheck - Convert untagged switch to tagged switch
//   - QF1011: staticcheck - Omit redundant type from variable declaration
//   - S1000:  staticcheck - Use plain channel send or receive instead of single-case select
//   - S1001:  staticcheck - Replace for loop with call to copy
//   - ST1005: staticcheck - Incorrectly formatted error string
//   - ST1016: staticcheck - Use consistent method receiver names
//   - ST1023: staticcheck - Redundant type in variable declaration
package main

import (
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	// standard go vet checks
	checks := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		slog.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		testinggoroutine.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}

	// exit checker, ineffassign, bodyclose
	checks = append(checks, ExitCheckAnalyzer, ineffassign.Analyzer, bodyclose.Analyzer)

	staticChecks := map[string]bool{
		"QF1002": true, //	Convert untagged switch to tagged switch
		"QF1011": true, // Omit redundant type from variable declaration
		"S1000":  true, //Use plain channel send or receive instead of single-case select
		"S1001":  true, //	Replace for loop with call to copy
		"ST1005": true, //	Incorrectly formatted error string
		"ST1016": true, //	Use consistent method receiver names
		"ST1023": true, //	Redundant type in variable declaration
	}

	// add analyzers from staticcheck
	for _, v := range staticcheck.Analyzers {
		// добавляем в массив нужные проверки
		switch {
		case v.Analyzer.Name[0:2] == "SA":
			checks = append(checks, v.Analyzer)
		case staticChecks[v.Analyzer.Name]:
			checks = append(checks, v.Analyzer)
		}
	}

	multichecker.Main(checks...)
}
