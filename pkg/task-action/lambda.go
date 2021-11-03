package action

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
	"github.com/suzuki-shunsuke/cmdx/pkg/util"
)

func awsSession(task *domain.Task) (*session.Session, error) {
	awsConfig := aws.Config{}
	if task.Region != "" {
		awsConfig.Region = aws.String(task.Region)
	}
	return session.NewSessionWithOptions(session.Options{
		Profile: task.Profile,
		Config:  awsConfig,
	})
}

func lambdaAction(ctx context.Context, task *domain.Task, vars map[string]interface{}) error {
	payload := make(map[string]interface{}, len(task.Payload))
	for _, payloadParam := range task.Payload {
		value, err := util.RenderTemplate(payloadParam.Value, vars)
		if err != nil {
			return fmt.Errorf("failed to parse the payload %s: %w", payloadParam.Name, err)
		}
		payload[payloadParam.Name] = value
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	sess, err := awsSession(task)
	if err != nil {
		return err
	}
	lambdaClient := lambda.New(sess)
	input := &lambda.InvokeInput{
		FunctionName: aws.String(task.FunctionName),
		Payload:      b,
	}
	if task.InvocationType != "" {
		input.InvocationType = aws.String(task.InvocationType)
	}
	if task.LogType != "" {
		input.LogType = aws.String(task.LogType)
	}
	out, err := lambdaClient.InvokeWithContext(ctx, input)
	if err != nil {
		return err
	}
	if out.Payload != nil {
		fmt.Println(string(out.Payload))
	}
	if task.LogType != "" && out.LogResult != nil {
		decoded, err := base64.StdEncoding.DecodeString(aws.StringValue(out.LogResult))
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, string(decoded))
	}
	return nil
}
