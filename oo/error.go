package oo

import (
	"errors"
	"fmt"
)

const (
	ESUCC         = "ESUCC"
	EBADGATEWAY   = "EBADGATEWAY"
	ENOTAUTH      = "ENOTAUTH"
	ENOTPERM      = "ENOTPERM"
	EPARAM        = "EPARAM"
	ESERVER       = "ESERVER"
	EFATAL        = "EFATAL"
	EEXISTS       = "EEXISTS"
	ENEXISTS      = "ENEXISTS"
	ETIMEOUT      = "ETIMEOUT"
	ENEEDCODE     = "ENEEDCODE"
	EPASSWD       = "EPASSWD"
	ETIMENOTALLOW = "ETIMENOTALLOW"
	EBALANCE      = "EBALANCE"
	ELIMITED      = "ELIMITED"
	ENOTALLOW     = "ENOTALLOW"
	ENODATA       = "ENODATA"
	UNSUPPORTED   = "UNSUPPORTED"
)

func ErrStr(eno string) string {
	switch eno {
	case ESUCC:
		return "Operation is successful"
	case EBADGATEWAY:
		return "Backen server has down"
	case ENOTAUTH:
		return "User not logged in or login has expired"
	case ENOTPERM:
		return "Permission denied"
	case EPARAM:
		return "Wrong parameter"
	case ESERVER:
		return "Internal server error"
	case EFATAL:
		return "Server fatal error"
	case EEXISTS:
		return "Unexpected existence"
	case ENEXISTS:
		return "Unexpected not existence"
	case ETIMEOUT:
		return "Access timeout"
	case ENEEDCODE:
		return "Need to provide a graphic verification code"
	case EPASSWD:
		return "Wrong password"
	case ETIMENOTALLOW:
		return "Operate during periods of time that are not allowed"
	case EBALANCE:
		return "Insufficient balance"
	case ELIMITED:
		return "The upper limit has been reached"
	case ENOTALLOW:
		return "Not allow to do that now"
	case ENODATA:
		return "No data"
	case UNSUPPORTED:
		return "Unsupported invocation"
	default:
		return eno
	}
}

type Error struct {
	Eno string `json:"eno,omitempty"`
	Str string `json:"str,omitempty"`
}

func (e *Error) Error() string {
	return e.Eno + ":" + ErrStr(e.Eno) + "; " + e.Str
}
func (e *Error) Errno() string {
	return e.Eno
}

func NewErrno(eno string, format ...interface{}) (e *Error) {
	var Str string
	if len(format) > 0 {
		ff, _ := format[0].(string)
		Str = errors.New(fmt.Sprintf(ff, format[1:]...)).Error()
	} else {
		Str = ErrStr(eno)
	}
	e = &Error{Eno: eno, Str: Str}
	return
}
func NewError(format ...interface{}) (e *Error) {
	return NewErrno(ESERVER, format...)
}
