// The codes this module refuses a caller with.
//
// A refusal a person or a client can see is a registered code, never a bare
// English sentence. The code is what a browser half switches on to choose a
// translation, and the English sentence travels beside it so a log line and an
// untranslated screen still read.
//
// Registering a code takes its documentation with it: what went wrong, what the
// caller should do, whether sending the same request again can work, and the
// parameters the sentence names. The sdk refuses a registration without them.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prusa3d/prusa-connect/sdk/fault"
)

// faultDomain is the first half of every code below. It is this module's
// namespace, which the host derived from the Go module path github.com/singlergold2/maintenance, so no
// other author can register a code under it. The sdk refuses the core
// namespace, which is what keeps this module's codes apart from the server's.
//
//nolint:gochecknoglobals // a domain built once at start-up; a const cannot hold one.
var faultDomain = mustDomain("com.github.aux.singlergold2.maintenance")

// The codes this module raises. Add a code here, and nowhere else, so the list
// a client branches on is one file.
//
//nolint:gochecknoglobals // definitions built once at start-up.
var (
	// RouteNotFound is a path this module does not serve.
	RouteNotFound = fault.Register(faultDomain, "ROUTE_NOT_FOUND", http.StatusNotFound, fault.Document(
		"This module does not serve that path",
		"Read openapi.json. Every path in it is relative to the mount the operator gave this module.",
		fault.RetryNever,
		fault.ShowDetail,
		"path",
	))

	// InvalidInput is a request this module cannot read.
	InvalidInput = fault.Register(faultDomain, "INVALID_INPUT", http.StatusBadRequest, fault.Document(
		"The request does not name a task this module can keep",
		"Send a task with a device, a name, and either interval_hours or interval_days.",
		fault.RetryNever,
		fault.ShowDetail,
		"kind",
	))

	// NoTask is an id no task carries.
	NoTask = fault.Register(faultDomain, "NO_TASK", http.StatusNotFound, fault.Document(
		"No task carries that id",
		"Read GET /overview and use the id it lists.",
		fault.RetryNever,
		fault.ShowDetail,
		"id",
	))

	// NoDatabase is an instance the operator granted no database.
	NoDatabase = fault.Register(faultDomain, "NO_DATABASE", http.StatusServiceUnavailable, fault.Document(
		"This instance has nowhere to keep its tasks",
		"Approve the module's database request and install it again.",
		fault.RetryNever,
		fault.ShowDetail,
	))

	// StoreFailed is the module's own database refusing a statement.
	StoreFailed = fault.Register(faultDomain, "STORE_FAILED", http.StatusInternalServerError, fault.Document(
		"The module's database refused the statement",
		"Read the reason beside this code and the server log.",
		fault.RetryAllowed,
		fault.ShowDetail,
		"reason",
	))

	// BuiltinTask is a task the module ships, which an operator may change but
	// not remove.
	BuiltinTask = fault.Register(faultDomain, "BUILTIN_TASK", http.StatusConflict, fault.Document(
		"This task ships with the module and cannot be removed",
		"Change its interval instead, or leave it in place.",
		fault.RetryNever,
		fault.ShowDetail,
	))

	// HistoryFailed is the print history being unreadable.
	HistoryFailed = fault.Register(faultDomain, "HISTORY_FAILED", http.StatusBadGateway, fault.Document(
		"The print history could not be read",
		"Read the reason beside this code. The host may be restarting.",
		fault.RetryAllowed,
		fault.ShowDetail,
		"reason",
	))
)

// faultString is one text parameter of a refusal.
func faultString(key, value string) fault.Param {
	return fault.StringParam(key, value)
}

// mustDomain builds this module's fault domain. A name the sdk refuses is a
// mistake in this file, so it stops the process at start-up rather than
// answering a caller with an empty code.
func mustDomain(name string) fault.Domain {
	domain, err := fault.NewDomain(name)
	if err != nil {
		panic(fmt.Sprintf("hello: %s is not a fault domain: %v", name, err))
	}

	return domain
}

// refuse writes one refusal as RFC 9457 frames it, which is the shape the
// generated client reads and the shape describeFault reads in the browser half:
// the code, the English sentence, the values that sentence names, and where the
// code is documented.
func refuse(w http.ResponseWriter, r *http.Request, raised *fault.Fault) {
	problem := raised.ProblemJSON(r.URL.Path)

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)

	_ = json.NewEncoder(w).Encode(problem)
}
