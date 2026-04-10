package app

import (
	"jarvis/internal/config"
	"jarvis/internal/logger"
	"jarvis/internal/output"
)

type State struct {
	Config     config.Config
	ConfigPath string
	Printer    *output.Printer
	Logger     *logger.Logger
	JSON       bool
	NoColor    bool
}
