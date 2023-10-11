package lambda

import (
	"encoding/json"
	authJwt "learn-to-code/internal/infrastructure/authentication/jwt"

	"github.com/aws/aws-lambda-go/events"
)

type RequestValidator struct {
	nextJsSecretParser *NextJsSecretParser
	jwtTokenValidator  *authJwt.Validator
}

func NewRequestValidator(nextJsSecretParser *NextJsSecretParser, jwtTokenValidator *authJwt.Validator) *RequestValidator {
	return &RequestValidator{
		nextJsSecretParser: nextJsSecretParser,
		jwtTokenValidator:  jwtTokenValidator,
	}
}

type Body struct {
	Input string `json:"input"`
}

func (r RequestValidator) ValidateRequest(request events.APIGatewayProxyRequest) (Body, string, error) {
	jwtToken, err := r.nextJsSecretParser.GetJwtTokenFromRequest(request)
	if err != nil {
		return Body{}, "", err
	}

	userID, err := r.jwtTokenValidator.ValidateAndGetUserID(jwtToken)
	if err != nil {
		return Body{}, "", err
	}

	var body Body
	err = json.Unmarshal([]byte(request.Body), &body)
	if err != nil {
		return Body{}, "", err
	}

	return body, userID, nil
}