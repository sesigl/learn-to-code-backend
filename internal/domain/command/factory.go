package command

import (
	"learn-to-code/internal/domain/command/data"
	"time"
)

type Factory struct {
}

func NewCommandFactory() *Factory {
	return &Factory{}
}

func (f *Factory) CreateStartQuizCommand(quizID string) Command {
	return NewCommand(data.StartQuizCommandType, data.NewStartQuizData(quizID), time.Now())
}