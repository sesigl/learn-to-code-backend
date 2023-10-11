package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"learn-to-code/internal/domain/eventsource"
	"learn-to-code/internal/domain/quiz/participant"
	"learn-to-code/internal/domain/quiz/participant/event"
	"learn-to-code/internal/infrastructure/config"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

type MarshalFunc func(v interface{}) ([]byte, error)
type UnmarshalFunc func(data []byte, v interface{}) error

type EventPo struct {
	AggregateID string    `dynamodbav:"aggregate_id"`
	Type        string    `dynamodbav:"type"`
	Version     uint      `dynamodbav:"version"`
	Payload     string    `dynamodbav:"payload"`
	CreatedAt   time.Time `dynamodbav:"created_at"`
}

type Repository struct {
	dbClient     *dynamodb.Client
	ctx          context.Context
	serializer   MarshalFunc
	deserializer UnmarshalFunc
	tableName    string
}

func NewDynamoDbParticipantRepository(ctx context.Context, environment config.Environment, client *dynamodb.Client) *Repository {

	tableName := fmt.Sprintf("%s_events", environment)

	return &Repository{
		dbClient:     client,
		ctx:          ctx,
		serializer:   json.Marshal,
		deserializer: json.Unmarshal,
		tableName:    tableName,
	}
}

func (r Repository) AppendEvents(participantID string, events []eventsource.Event) error {
	for _, e := range events {
		err := r.appendEvent(participantID, e)

		if err != nil {
			return err
		}
	}

	return nil
}

func (r Repository) appendEvent(participantID string, e eventsource.Event) error {
	serializedEvent, err := r.serializer(e)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: &r.tableName,
		Item: map[string]types.AttributeValue{
			"aggregate_id": &types.AttributeValueMemberS{Value: participantID},
			"version":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", e.GetVersion())},
			"type":         &types.AttributeValueMemberS{Value: reflect.TypeOf(e).Name()},
			"payload":      &types.AttributeValueMemberS{Value: string(serializedEvent)},
			"created_at":   &types.AttributeValueMemberS{Value: e.GetCreatedAt().Format(time.RFC3339)},
		},
	}
	_, err = r.dbClient.PutItem(r.ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (r Repository) FindByID(id string) (participant.Participant, error) {
	input := &dynamodb.QueryInput{
		TableName: &r.tableName,
		KeyConditions: map[string]types.Condition{
			"aggregate_id": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: id},
				},
			},
		},
	}

	output, err := r.dbClient.Query(r.ctx, input)
	if err != nil {
		return participant.Participant{}, err
	}

	if len(output.Items) == 0 {
		return participant.Participant{}, participant.ErrNotFound
	}

	var events []eventsource.Event

	for _, outputItem := range output.Items {

		eventPo := EventPo{}
		err := attributevalue.UnmarshalMap(outputItem, &eventPo)
		if err != nil {
			return participant.Participant{}, err
		}

		switch eventPo.Type {

		case event.ParticipantCreatedTypeName:
			joinedQuizEvent := &event.ParticipantCreated{}

			err := r.deserializer([]byte(eventPo.Payload), joinedQuizEvent)
			if err != nil {
				return participant.Participant{}, err
			}

			events = append(events, *joinedQuizEvent)
		case event.FinishedQuizTypeName:
			finishedQuiz := &event.FinishedQuiz{}

			err := r.deserializer([]byte(eventPo.Payload), finishedQuiz)
			if err != nil {
				return participant.Participant{}, err
			}

			events = append(events, *finishedQuiz)

		case event.StartedQuizTypeName:
			startedQuiz := &event.StartedQuiz{}

			err := r.deserializer([]byte(eventPo.Payload), startedQuiz)
			if err != nil {
				return participant.Participant{}, err
			}

			events = append(events, *startedQuiz)

		default:
			panic(fmt.Errorf("unknown type '%s' while reading persisted events", eventPo.Type))
		}

	}

	p, newFromEventsErr := participant.NewFromEvents(events)

	return p, newFromEventsErr
}