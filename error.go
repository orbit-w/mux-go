package mux

import (
	"errors"
	"fmt"
	"strings"
)

/*
   @Author: orbit-w
   @File: error
   @2024 4月 周日 23:32
*/

var (
	ErrCancel             = errors.New("context canceled")
	ErrConnDone           = errors.New("error_the_conn_is_done")
	ErrVirtualConnUpLimit = errors.New("error_virtual_conn_up_limit")
)

func isErrCanceled(err error) bool {
	return err != nil && strings.Contains(err.Error(), "context canceled")
}

func newStreamBufSetErr(err error) error {
	return errors.New(fmt.Sprintf("NewStream set failed: %s", err.Error()))
}

func newDecodeErr(err error) error {
	return errors.New(fmt.Sprintf("decode data failed: %s", err.Error()))
}
