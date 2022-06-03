package main

import (
	"errors"
)

// self defined Errors
var (
	InvalidCluster     = errors.New("invalid cluster")
	FindIndexNotFound  = errors.New("findIndex does not find pattern")
	ParseMessageError  = errors.New("parse message error")
	ConvertWrongType   = errors.New("parse result convert to type fail")
	ParseSplitError    = errors.New("split message fail")
	ResultInvalid      = errors.New("invalid Result")
	NoPingResultFound  = errors.New("no Ping Result")
	NoPingResultRecord = errors.New("no Ping Result Record")
	NoPingResultShort  = errors.New("PingResultError has no shortname")
	TransactionLoss    = errors.New("TransactionLoss")
)
