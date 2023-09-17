package logger

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Hook struct {
	Ctx       *context.Context
	LogLevels []log.Level
}

func (hook *Hook) Fire(entry *log.Entry) error {

	if *hook.Ctx == nil {
		return nil
	}

	caller := fmt.Sprintf("File: %s\nLine: %d\nFunction:%s\n", entry.Caller.File, entry.Caller.Line, entry.Caller.Function)
	footer := "Please report this issue at https://github.com/ananthakumaran/paisa/issues."
	message := fmt.Sprintf("Paisa has encountered a critical error and needs to close.\n\n %s\n\n%s\n%s", entry.Message, caller, footer)

	_, err := runtime.MessageDialog(*hook.Ctx, runtime.MessageDialogOptions{
		Type:          runtime.ErrorDialog,
		Title:         "Critical Error",
		Message:       message,
		DefaultButton: "Ok",
	})

	return err
}

func (hook *Hook) Levels() []log.Level {
	return hook.LogLevels
}

type Logger struct {
}

func (logger *Logger) Print(message string) {
	log.Print(message)
}

func (logger *Logger) Trace(message string) {
	log.Trace(message)
}

func (logger *Logger) Debug(message string) {
	log.Debug(message)
}

func (logger *Logger) Info(message string) {
	log.Info(message)
}

func (logger *Logger) Warning(message string) {
	log.Warning(message)
}

func (logger *Logger) Error(message string) {
	log.Error(message)
}

func (logger *Logger) Fatal(message string) {
	log.Fatal(message)
}
