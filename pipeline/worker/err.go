package worker

import "errors"

var ErrTaskChClosed = errors.New("pipeline task channal is closed")

var ErrTaskChNotReady = errors.New("pipeline task channal is not ready")

var ErrTaskBusy = errors.New("task busy")

var ErrContentType = errors.New("only supports []byte and string type data")

var ErrResultDropped = errors.New("result dropped")
