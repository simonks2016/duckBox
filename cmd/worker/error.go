package worker

type TransientError struct{ Err error } // 可重试
func (e TransientError) Error() string  { return e.Err.Error() }
func Transient(err error) error         { return TransientError{Err: err} }

type PermanentError struct{ Err error } // 不重试，直接ACK（可旁路到 reject.*）
func (e PermanentError) Error() string  { return e.Err.Error() }
func Permanent(err error) error         { return PermanentError{Err: err} }

type DropToDLQ struct{ Reason string } // 直接进DLQ（原样 body）
func (e DropToDLQ) Error() string      { return "drop_to_dlq: " + e.Reason }
func Drop(reason string) error         { return DropToDLQ{Reason: reason} }
