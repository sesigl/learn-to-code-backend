package event

import (
	"learn-to-code/internal/domain/eventsource"
	"reflect"
)

type StartedQuiz struct {
	QuizID                    string
	RequiredQuestionsAnswered []string
	eventsource.EventBase
}

var StartedQuizTypeName = reflect.TypeOf(StartedQuiz{}).Name()
