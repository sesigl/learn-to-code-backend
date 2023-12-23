package application

import (
	"learn-to-code/internal/domain/quiz/participant"
	"learn-to-code/internal/domain/quiz/participant/command"
	"learn-to-code/internal/domain/quiz/participant/projection"
)

type ParticipantApplicationService struct {
	participantRepository  participant.Repository
	startQuizToEventMapper *command.ParticipantCommandApplier
}

func NewPartcipantApplicationService(participantRepository participant.Repository, participantCommandApplier *command.ParticipantCommandApplier) *ParticipantApplicationService {
	return &ParticipantApplicationService{
		participantRepository:  participantRepository,
		startQuizToEventMapper: participantCommandApplier,
	}
}

func (as *ParticipantApplicationService) GetStartedQuizCount(participantID string) (int, error) {
	p, err := as.participantRepository.FindOrCreateByID(participantID)
	if err != nil {
		return 0, err
	}

	return p.GetStartedQuizCount(), nil
}

func (as *ParticipantApplicationService) ProcessCommand(command command.Command, participantID string) error {

	p, err := as.participantRepository.FindOrCreateByID(participantID)
	if err != nil {
		return err
	}

	p, err = as.startQuizToEventMapper.ApplyCommand(command, p)
	if err != nil {
		return err
	}

	appendEventErr := as.participantRepository.StoreEvents(p.GetID(), p.GetNewEventsAndUpdatePersistedVersion())

	return appendEventErr
}

func (as *ParticipantApplicationService) GetQuizzes(participantID string) (projection.QuizOverview, error) {
	p, err := as.participantRepository.FindOrCreateByID(participantID)
	if err != nil {
		return projection.QuizOverview{}, err
	}

	quizOverview := projection.NewQuizOverview(p)

	return quizOverview, nil
}

func (as *ParticipantApplicationService) GetQuizAttemptDetail(participantID string, quizId string, attemptId int) (projection.QuizAttemptDetail, error) {
	p, err := as.participantRepository.FindOrCreateByID(participantID)
	if err != nil {
		return projection.QuizAttemptDetail{}, err
	}

	quizOverview, err := projection.NewQuizAttemptDetail(p, quizId, attemptId)

	return quizOverview, err
}
