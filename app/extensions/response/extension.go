package response

import "github.com/thrisp/flotilla/extension"

func mkFunction(k string, v interface{}) extension.Function {
	return extension.NewFunction(k, v)
}

var responseFns = []extension.Function{
	mkFunction("abort", abort),
	mkFunction("header_now", headerNow),
	mkFunction("header_write", headerWrite),
	mkFunction("header_modify", headerModify),
	mkFunction("is_written", isWritten),
	mkFunction("redirect", redirect),
	mkFunction("serve_file", serveFile),
	mkFunction("serve_plain", servePlain),
	mkFunction("write_to_response", writeToResponse),
}

var Extension extension.Extension

func init() {
	Extension = extension.New("Response_Extension", responseFns...)
}
