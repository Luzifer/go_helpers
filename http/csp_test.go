package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCSPClone(t *testing.T) {
	c1 := CSP{"default-src": []CSPSourceValue{CSPSrcNone}}
	c2 := c1.Clone()

	c2.Add("default-src", CSPSrcSelf) // Makes no sense in real world!

	assert.NotEqual(t, c1["default-src"], c2["default-src"], "expect c1 not to have changed")
}

func TestCSPToHeaderValue(t *testing.T) {
	c := CSP{}
	c.Set("default-src", CSPSrcNone)

	c.Set("connect-src", CSPSrcSelf)
	c.Set("font-src", CSPSrcSelf)
	c.Set("img-src", CSPSrcSelf)
	c.Add("img-src", CSPSrcSchemeData)
	c.Set("script-src", CSPSrcSelf)
	c.Add("script-src", CSPSrcUnsafeInline)
	c.Set("style-src", CSPSrcSelf)

	assert.Equal(t,
		"connect-src 'self';default-src 'none';font-src 'self';img-src 'self' data:;script-src 'self' 'unsafe-inline';style-src 'self'",
		c.ToHeaderValue())
}
