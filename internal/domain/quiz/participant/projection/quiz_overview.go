package projection

import (
	"fmt"
	"learn-to-code/internal/domain/quiz/participant"
	"learn-to-code/internal/domain/quiz/participant/event"
	"learn-to-code/internal/domain/quiz/participant/projection/quizattemptdetail"
)

type QuizOverview struct {
	ActiveQuizzes   map[string][]QuizAttemptOverview
	FinishedQuizzes map[string][]QuizAttemptOverview
}

func (qo QuizOverview) GetFinishedQuizLatestAttempt(quizID string) (QuizAttemptOverview, error) {
	finishedQuiz, ok := qo.FinishedQuizzes[quizID]
	if !ok {
		return QuizAttemptOverview{}, fmt.Errorf("quiz %s does not exist", quizID)
	}

	if len(qo.FinishedQuizzes[quizID]) == 0 {
		return QuizAttemptOverview{}, fmt.Errorf("no quiz with id %s finished yet", quizID)
	}

	latestQuizAttempt := finishedQuiz[len(finishedQuiz)-1]

	return latestQuizAttempt, nil
}

func NewQuizOverview(p participant.Participant) (QuizOverview, error) {

	qo := QuizOverview{
		ActiveQuizzes:   map[string][]QuizAttemptOverview{},
		FinishedQuizzes: map[string][]QuizAttemptOverview{},
	}

	quizAttemptCounter := map[string]int{}
	quizFinishCounter := map[string]int{}
	quizAttemptAnswers := map[string]map[string]bool{}

	var activeQuizAttempts = map[string]*QuizAttemptOverview{}

	for _, generalEvent := range p.GetEvents() {

		switch e := generalEvent.(type) {

		case event.StartedQuiz:
			_, ok := quizAttemptCounter[e.QuizID]
			if !ok {
				quizAttemptCounter[e.QuizID] = 0
			}
			quizAttemptCounter[e.QuizID]++

			if activeQuizAttempts[e.QuizID] != nil {
				return QuizOverview{}, fmt.Errorf("invalid multiple active attempts for quiz %s", e.QuizID)
			}

			activeQuizAttempts[e.QuizID] = &QuizAttemptOverview{
				QuizID:              e.QuizID,
				AttemptID:           quizAttemptCounter[e.QuizID],
				QuestionsWithAnswer: map[string]string{},
			}

		case event.SelectedAnswer:
			attemptAnswerCounterIndex := getQuizAttemptCorrectAnswerCounterIndex(e.QuizID, quizAttemptCounter)
			_, ok := quizAttemptAnswers[attemptAnswerCounterIndex]
			if !ok {
				quizAttemptAnswers[attemptAnswerCounterIndex] = map[string]bool{}
			}

			if e.IsCorrect {
				quizAttemptAnswers[attemptAnswerCounterIndex][e.QuestionID] = true
			} else {
				quizAttemptAnswers[attemptAnswerCounterIndex][e.QuestionID] = false
			}

		case event.FinishedQuiz:
			_, ok := quizFinishCounter[e.QuizID]
			if !ok {
				quizFinishCounter[e.QuizID] = 0
			}
			quizFinishCounter[e.QuizID]++

			_, ok = qo.FinishedQuizzes[e.QuizID]
			if !ok {
				qo.FinishedQuizzes[e.QuizID] = []QuizAttemptOverview{}
			}

			quizAttemptCorrectAnswerCounter, quizAttemptWrongAnswerCounter := getQuizAnswerCounters(e.QuizID, quizAttemptCounter, quizAttemptAnswers)

			activeQuizAttempts[e.QuizID].Pass = isPass(
				quizAttemptCorrectAnswerCounter,
				quizAttemptWrongAnswerCounter,
			)

			activeQuizAttempts[e.QuizID].QuestionCorrectRatio = getCorrectnessRatio(
				quizAttemptCorrectAnswerCounter,
				quizAttemptWrongAnswerCounter,
			)

			qo.FinishedQuizzes[e.QuizID] = append(qo.FinishedQuizzes[e.QuizID], *activeQuizAttempts[e.QuizID])
			delete(activeQuizAttempts, e.QuizID)
		}
	}

	for activeQuizID, activeQuizAttemptOverview := range activeQuizAttempts {
		_, ok := qo.ActiveQuizzes[activeQuizID]
		if !ok {
			qo.ActiveQuizzes[activeQuizID] = []QuizAttemptOverview{}
		}

		qo.ActiveQuizzes[activeQuizID] = append(qo.ActiveQuizzes[activeQuizID], QuizAttemptOverview{
			QuizID:              activeQuizID,
			AttemptID:           activeQuizAttemptOverview.AttemptID,
			QuestionsWithAnswer: activeQuizAttemptOverview.QuestionsWithAnswer,
			Pass:                false,
		})
	}

	return qo, nil
}

func getQuizAnswerCounters(quizID string, quizAttemptCounter map[string]int, quizAttemptAnswers map[string]map[string]bool) (int, int) {
	attemptAnswerCounterIndex := getQuizAttemptCorrectAnswerCounterIndex(quizID, quizAttemptCounter)

	quizAttemptCorrectAnswerCounter := 0
	quizAttemptWrongAnswerCounter := 0

	for _, correct := range quizAttemptAnswers[attemptAnswerCounterIndex] {
		if correct {
			quizAttemptCorrectAnswerCounter++
		} else {
			quizAttemptWrongAnswerCounter++
		}
	}
	return quizAttemptCorrectAnswerCounter, quizAttemptWrongAnswerCounter
}

func getQuizAttemptCorrectAnswerCounterIndex(quizID string, quizAttemptCounter map[string]int) string {
	return quizID + "-" + fmt.Sprintf("%d", quizAttemptCounter[quizID])
}

func getCorrectnessRatio(correctAnswerCounter int, incorrectAnswerCounter int) float64 {
	if correctAnswerCounter+incorrectAnswerCounter == 0 {
		return 0
	}

	return (float64(correctAnswerCounter) / (float64(correctAnswerCounter) + float64(incorrectAnswerCounter)))
}

func isPass(correctAnswerCounter int, incorrectAnswerCounter int) bool {
	if correctAnswerCounter+incorrectAnswerCounter == 0 {
		return true
	}

	return getCorrectnessRatio(correctAnswerCounter, incorrectAnswerCounter) >= quizattemptdetail.QuizPassThresold
}
