package worker

import "errors"

var ErrTaskChClosed = errors.New("pipeline task channal is closed")

var ErrTaskChNotReady = errors.New("pipeline task channal is not ready")

var ErrScriptExists = errors.New("the current pipeline script exists")
