package cookie

import "github.com/flxtilla/cxre/extension"

func mkFunction(k string, v interface{}) extension.Function {
	return extension.NewFunction(k, v)
}

var cookieFns = []extension.Function{
	mkFunction("cookie", cookie),
	mkFunction("securecookie", securecookie),
	mkFunction("cookies", cookies),
	mkFunction("readcookies", readcookies),
}

var Extension extension.Extension

func init() {
	Extension = extension.New("cookie_fxtension", cookieFns...)
}
