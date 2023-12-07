package projection_test

import (
	"learn-to-code/internal/domain/quiz/participant"
	"learn-to-code/internal/domain/quiz/participant/projection"
	"learn-to-code/internal/infrastructure/go/util/err"
	"learn-to-code/internal/infrastructure/go/util/uuid"
	"learn-to-code/internal/infrastructure/inmemory"
	"testing"
)

func TestNewQuizOverview_EmptyForEmptyUsers(t *testing.T) {

	p := newParticipant()

	qo := projection.NewQuizOverview(p)

	if len(qo.ActiveQuizzes) != 0 {
		t.Fatalf("empty user has active quizzes")
	}

	if len(qo.FinishedQuizzes) != 0 {
		t.Fatalf("empty user has finished quizzes")
	}
}

func TestNewQuizOverview_AddsActiveQuizzes(t *testing.T) {

	p := newParticipant()

	activeQuizId := inmemory.QuizID
	finishedQuizId := "2d107555-e311-4a52-a5f9-6997e88c407c"

	p.StartQuiz(activeQuizId, nil)

	p.StartQuiz(finishedQuizId, nil)
	p.FinishQuiz(finishedQuizId)

	qo := projection.NewQuizOverview(p)

	if len(qo.ActiveQuizzes) != 1 {
		t.Fatalf("user with active quiz has no active quiz in overview")
	}

	if len(qo.FinishedQuizzes) != 1 {
		t.Fatalf("user with finished quiz has no finished quiz in overview")
	}
}

func newParticipant() participant.Participant {
	return err.PanicIfError1(participant.NewParticipant(uuid.MustNewRandomAsString()))
}