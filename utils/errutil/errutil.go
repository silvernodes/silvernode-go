package errutil

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
)

var _EOF error

func init() {
	_EOF = errors.New("EOF")
}

func EOF() error {
	return _EOF
}

func IsEOF(err error) bool {
	return err.Error() == "EOF"
}

type TagError struct {
	tag  string
	text string
}

func NewWithTag(tag string, text string) error {
	err := new(TagError)
	err.tag = tag
	err.text = text
	return err
}

func (err *TagError) Tag() string {
	return err.tag
}

func (err *TagError) Error() string {
	return "[" + err.tag + "]" + err.text
}

type CodeError struct {
	code int
	text string
}

func NewWithCode(code int, text string) error {
	err := new(CodeError)
	err.code = code
	err.text = text
	return err
}

func (err *CodeError) Code() int {
	return err.code
}

func (err *CodeError) Error() string {
	return "[" + strconv.Itoa(err.code) + "]" + err.text
}

func New(text string) error {
	return errors.New(text)
}

func Extend(text string, err error) error {
	return errors.New(text + ":" + err.Error())
}

func ExtendWithCode(code int, text string, err error) error {
	return NewWithCode(code, text+":"+err.Error())
}

func NewAndPrint(text string) error {
	fmt.Println(text)
	return errors.New(text)
}

func Try(do func(), catch func(error)) {
	if do == nil {
		return
	}
	defer Catch(catch)
	do()
}

func Catch(catch func(error)) {
	if err := recover(); err != nil {
		errIns := errors.New(
			" :: Try-Catch :: \r" +
				fmt.Sprint(err) +
				"\r=============== - CallStackInfo - =============== \r" +
				string(debug.Stack()))
		if catch == nil {
			ReportError(errIns)
			return
		}
		catch(errIns)
	}
}

var _customErr func(error) = nil

func CustomErrFunc(custom func(error)) {
	_customErr = custom
}

func ReportError(err error) {
	if _customErr != nil {
		_customErr(err)
	} else {
		fmt.Println("[ERROR]::" + err.Error())
	}
}
