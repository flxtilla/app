package session

import "github.com/thrisp/flotilla/extension"

func mkFunction(k string, v interface{}) extension.Function {
	return extension.NewFunction(k, v)
}

var sessionFns = []extension.Function{
	mkFunction("delete_session", deleteSession),
	mkFunction("get_session", getSession),
	mkFunction("session", returnSession),
	mkFunction("set_session", setSession),
}

var Extension extension.Extension = extension.New("Session_Extension", sessionFns...)
