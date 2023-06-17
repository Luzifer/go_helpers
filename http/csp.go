package http

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
)

type (
	// CSP is a non-concurrency-safe map to hold a Content-Security-Policy
	// and manipulate it afterwards to eventually render it into its
	// header-representation
	CSP map[string][]CSPSourceValue
	// CSPHashAlgo defines the available hash algorithms
	CSPHashAlgo string
	// CSPSourceValue represents an value in the map for a given directive
	CSPSourceValue string
)

// Collection of pre-defined values. For documentation see
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/Sources
const (
	CSPHashSHA256 CSPHashAlgo = "sha256"
	CSPHashSHA384 CSPHashAlgo = "sha384"
	CSPHashSHA512 CSPHashAlgo = "sha512"

	CSPSrcNone           CSPSourceValue = "'none'"
	CSPSrcReportSample   CSPSourceValue = "'report-sample'"
	CSPSrcSelf           CSPSourceValue = "'self'"
	CSPSrcStrictDynamic  CSPSourceValue = "'strict-dynamic'"
	CSPSrcUnsafeEval     CSPSourceValue = "'unsafe-eval'"
	CSPSrcUnsafeHashes   CSPSourceValue = "'unsafe-hashes'"
	CSPSrcUnsafeInline   CSPSourceValue = "'unsafe-inline'"
	CSPSrcWASMUnsafeEval CSPSourceValue = "'wasm-unsafe-eval'"

	CSPSrcSchemeData        CSPSourceValue = "data:"
	CSPSrcSchemeMediastream CSPSourceValue = "mediastream:"
	CSPSrcSchemeBlob        CSPSourceValue = "blob:"
	CSPSrcSchemeFilesystem  CSPSourceValue = "filesystem:"
)

// CSPSrcHash takes an algo (sha256, sha384 or sha512) and the sum
// value and converts it into the right representation for the header
func CSPSrcHash(algo CSPHashAlgo, sum []byte) CSPSourceValue {
	return CSPSourceValue(fmt.Sprintf("'%s-%s'", algo, base64.StdEncoding.EncodeToString(sum)))
}

// CSPSrcNonce takes a base64 encoded nonce value and converts it
// into the right representation for the header
func CSPSrcNonce(b64Value string) CSPSourceValue {
	return CSPSourceValue(fmt.Sprintf("'nonce-%s'", b64Value))
}

// Add adds a single CSPSourceValue to the given directive
func (c CSP) Add(directive string, value CSPSourceValue) {
	c[directive] = append(c[directive], value)
}

// Clone copies the CSP into a new one for manipulation
func (c CSP) Clone() CSP {
	n := make(CSP)
	for dir, vals := range c {
		n[dir] = append([]CSPSourceValue{}, vals...)
	}
	return n
}

// Set replaces all present values for the given directive
func (c CSP) Set(directive string, value CSPSourceValue) {
	c[directive] = []CSPSourceValue{value}
}

// ToHeaderValue encodes the policy into the format expected in the
// `Content-Security-Policy` header.
func (c CSP) ToHeaderValue() string {
	var keys []string
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		lst := []string{k}
		for _, v := range c[k] {
			lst = append(lst, string(v))
		}
		parts = append(parts, strings.Join(lst, " "))
	}

	return strings.Join(parts, ";")
}
