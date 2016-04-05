package api

import (
)

const (
	RunStateStart  = "running"
	RunStateStop   = "stopped"
	RunStatePause  = "suspended"
	RunStateKill   = "halted"
	RunStateBusy   = "busy"
	RunStateReset  = "reset"
)

func isOkStatus(code int) bool {
	codes := map[int]bool{
		200: true,
		201: true,
		204: true,
		401: false,
		404: false,
		409: false,
		422: false,
		423: false,
		429: false,
		500: false,
	}

	return codes[code]
}

func isBusy(code int) bool {
	return code == 423
}