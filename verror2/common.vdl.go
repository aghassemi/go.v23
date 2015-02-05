// This file was auto-generated by the veyron vdl tool.
// Source: common.vdl

package verror2

import (
	// VDL system imports
	"v.io/core/veyron2/context"
	"v.io/core/veyron2/i18n"
)

var (
	// Unknown means the error has no known ID.  A more specific error should
	// always be used, if possible.  Unknown is typically only used when
	// automatically converting errors that do not contain an ID.
	Unknown = Register("v.io/core/veyron2/verror.Unknown", NoRetry, "{1:}{2:} Error{:_}")
	// Internal means an internal error has occurred.  A more specific error
	// should always be used, if possible.
	Internal = Register("v.io/core/veyron2/verror.Internal", NoRetry, "{1:}{2:} Internal error{:_}")
	// EOF means the end-of-file has been reached; more generally, no more input
	// data is available.
	EOF = Register("v.io/core/veyron2/verror.EOF", NoRetry, "{1:}{2:} EOF{:_}")
	// BadArg means the arguments to an operation are invalid or incorrectly
	// formatted.
	BadArg = Register("v.io/core/veyron2/verror.BadArg", NoRetry, "{1:}{2:} Bad argument{:_}")
	// BadState means an operation was attempted on an object while the object was
	// in an incompatible state.
	BadState = Register("v.io/core/veyron2/verror.BadState", NoRetry, "{1:}{2:} Invalid state{:_}")
	// Exist means that the requested item already exists; typically returned when
	// an attempt to create an item fails because it already exists.
	Exist = Register("v.io/core/veyron2/verror.Exist", NoRetry, "{1:}{2:} Already exists{:_}")
	// NoExist means that the requested item does not exist; typically returned
	// when an attempt to lookup an item fails because it does not exist.
	NoExist = Register("v.io/core/veyron2/verror.NoExist", NoRetry, "{1:}{2:} Does not exist{:_}")
	// NoExistOrNoAccess means that either the requested item does not exist, or
	// is inaccessible.  Typically returned when the distinction between existence
	// and inaccessiblity needs to remain hidden, as a privacy feature.
	NoExistOrNoAccess = Register("v.io/core/veyron2/verror.NoExistOrNoAccess", NoRetry, "{1:}{2:} Does not exist or access denied{:_}")
	// The following errors can occur during the process of establishing
	// an RPC connection.
	// NoExist (see above) is returned if the name of the server fails to
	// resolve any addresses.
	// NoServers is returned when the servers returned for the supplied name
	// are somehow unusable or unreachable by the client.
	// NoAccess is returned when a server does not authorize a client.
	// NotTrusted is returned when a client does not trust a server.
	//
	// TODO(toddw): These errors and descriptions were added by Cos; consider
	// moving the IPC-related ones into the ipc package.
	NoServers        = Register("v.io/core/veyron2/verror.NoServers", RetryRefetch, "{1:}{2:} No usable servers found{:_}")
	NoAccess         = Register("v.io/core/veyron2/verror.NoAccess", RetryRefetch, "{1:}{2:} Access denied{:_}")
	NotTrusted       = Register("v.io/core/veyron2/verror.NotTrusted", RetryRefetch, "{1:}{2:} Client does not trust server{:_}")
	NoServersAndAuth = Register("v.io/core/veyron2/verror.NoServersAndAuth", RetryRefetch, "{1:}{2:} Has no usable servers and is either not trusted or access was denied{:_}")
	// Aborted means that an operation was not completed because it was aborted by
	// the receiver.  A more specific error should be used if it would help the
	// caller decide how to proceed.
	Aborted = Register("v.io/core/veyron2/verror.Aborted", NoRetry, "{1:}{2:} Aborted{:_}")
	// BadProtocol means that an operation was not completed because of a protocol
	// or codec error.
	BadProtocol = Register("v.io/core/veyron2/verror.BadProtocol", NoRetry, "{1:}{2:} Bad protocol or type{:_}")
	// Canceled means that an operation was not completed because it was
	// explicitly cancelled by the caller.
	Canceled = Register("v.io/core/veyron2/verror.Canceled", NoRetry, "{1:}{2:} Canceled{:_}")
	// Timeout means that an operation was not completed before the time deadline
	// for the operation.
	Timeout = Register("v.io/core/veyron2/verror.Timeout", NoRetry, "{1:}{2:} Timeout{:_}")
)

func init() {
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(Unknown.ID), "{1:}{2:} Error{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(Internal.ID), "{1:}{2:} Internal error{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(EOF.ID), "{1:}{2:} EOF{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(BadArg.ID), "{1:}{2:} Bad argument{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(BadState.ID), "{1:}{2:} Invalid state{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(Exist.ID), "{1:}{2:} Already exists{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(NoExist.ID), "{1:}{2:} Does not exist{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(NoExistOrNoAccess.ID), "{1:}{2:} Does not exist or access denied{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(NoServers.ID), "{1:}{2:} No usable servers found{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(NoAccess.ID), "{1:}{2:} Access denied{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(NotTrusted.ID), "{1:}{2:} Client does not trust server{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(NoServersAndAuth.ID), "{1:}{2:} Has no usable servers and is either not trusted or access was denied{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(Aborted.ID), "{1:}{2:} Aborted{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(BadProtocol.ID), "{1:}{2:} Bad protocol or type{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(Canceled.ID), "{1:}{2:} Canceled{:_}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(Timeout.ID), "{1:}{2:} Timeout{:_}")
}

// MakeUnknown returns an error with the Unknown ID.
func MakeUnknown(ctx *context.T) error {
	return Make(Unknown, ctx)
}

// MakeInternal returns an error with the Internal ID.
func MakeInternal(ctx *context.T) error {
	return Make(Internal, ctx)
}

// MakeEOF returns an error with the EOF ID.
func MakeEOF(ctx *context.T) error {
	return Make(EOF, ctx)
}

// MakeBadArg returns an error with the BadArg ID.
func MakeBadArg(ctx *context.T) error {
	return Make(BadArg, ctx)
}

// MakeBadState returns an error with the BadState ID.
func MakeBadState(ctx *context.T) error {
	return Make(BadState, ctx)
}

// MakeExist returns an error with the Exist ID.
func MakeExist(ctx *context.T) error {
	return Make(Exist, ctx)
}

// MakeNoExist returns an error with the NoExist ID.
func MakeNoExist(ctx *context.T) error {
	return Make(NoExist, ctx)
}

// MakeNoExistOrNoAccess returns an error with the NoExistOrNoAccess ID.
func MakeNoExistOrNoAccess(ctx *context.T) error {
	return Make(NoExistOrNoAccess, ctx)
}

// MakeNoServers returns an error with the NoServers ID.
func MakeNoServers(ctx *context.T) error {
	return Make(NoServers, ctx)
}

// MakeNoAccess returns an error with the NoAccess ID.
func MakeNoAccess(ctx *context.T) error {
	return Make(NoAccess, ctx)
}

// MakeNotTrusted returns an error with the NotTrusted ID.
func MakeNotTrusted(ctx *context.T) error {
	return Make(NotTrusted, ctx)
}

// MakeNoServersAndAuth returns an error with the NoServersAndAuth ID.
func MakeNoServersAndAuth(ctx *context.T) error {
	return Make(NoServersAndAuth, ctx)
}

// MakeAborted returns an error with the Aborted ID.
func MakeAborted(ctx *context.T) error {
	return Make(Aborted, ctx)
}

// MakeBadProtocol returns an error with the BadProtocol ID.
func MakeBadProtocol(ctx *context.T) error {
	return Make(BadProtocol, ctx)
}

// MakeCanceled returns an error with the Canceled ID.
func MakeCanceled(ctx *context.T) error {
	return Make(Canceled, ctx)
}

// MakeTimeout returns an error with the Timeout ID.
func MakeTimeout(ctx *context.T) error {
	return Make(Timeout, ctx)
}
