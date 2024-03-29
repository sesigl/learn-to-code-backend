package main

import (
	"fmt"
	"learn-to-code/internal/domain/command"
	"learn-to-code/internal/infrastructure/config"
	"learn-to-code/internal/infrastructure/local"
	"learn-to-code/internal/infrastructure/testing/json"
	"learn-to-code/internal/interfaces/lambda/participant"
	"learn-to-code/internal/interfaces/lambda/participant/quiz"
	"testing"
)

var eventBody = fmt.Sprintf(`
{
   "createdAt":"2023-11-17T04:55:24.059Z",
   "data": {
		"quizId":"fcf7890f-9c72-46d3-931e-34494307be37",
		"requiredQuestionsAnswered": ["f5b70d7d-3461-4cf7-978d-2b0caf77db1e"]
	},
   "type": "%s"
}
`, command.StartQuizCommandType)

func TestGetQuizOverview_Returns200(t *testing.T) {
	environmentCreator := local.NewEnvironmentCreator(config.Test)
	defer environmentCreator.Terminate()

	environmentCreator.ExecuteLambdaHandlerWithPostBody(participant.NewPostParticipantCommandHandler, eventBody)

	getOverviewResponse := environmentCreator.ExecuteLambdaHandler(quiz.NewGetParticipantQuizOverviewHandler)

	if getOverviewResponse.StatusCode != 200 {
		t.Fatalf("lambda did not succeed, status code: %v", getOverviewResponse.StatusCode)
	}

	activeQuizzesResponse := json.GetJSONPathValue(getOverviewResponse, "$.activeQuizzes")

	if len(activeQuizzesResponse.(map[string]interface{})) != 1 {
		t.Fatalf("no active quizzes in overview response")
	}
}

func TestGetQuizOverview_ReturnsPassInformation(t *testing.T) {
	environmentCreator := local.NewEnvironmentCreator(config.Test)
	defer environmentCreator.Terminate()

	environmentCreator.ExecuteLambdaHandlerWithPostBody(participant.NewPostParticipantCommandHandler, eventBody)

	getOverviewResponse := environmentCreator.ExecuteLambdaHandler(quiz.NewGetParticipantQuizOverviewHandler)

	if getOverviewResponse.StatusCode != 200 {
		t.Fatalf("lambda did not succeed, status code: %v", getOverviewResponse.StatusCode)
	}

	activeQuizzesResponse := json.GetJSONPathValue(getOverviewResponse, "$.activeQuizzes[\"fcf7890f-9c72-46d3-931e-34494307be37\"][0]")

	if activeQuizzesResponse.(map[string]interface{})["pass"] != false {
		t.Fatalf("expected active quiz to have pass false, but it is true")
	}

	if activeQuizzesResponse.(map[string]interface{})["questionCorrectRatio"] != 0.0 {
		t.Fatalf("expected active quiz to have questionCorrectRatio false, but it is true")
	}
}
