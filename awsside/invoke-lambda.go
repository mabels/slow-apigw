package awsside

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type FunctionWrapper struct {
	LambdaClient *lambda.Client
	function     string
}

func (wrapper FunctionWrapper) Invoke(ctx context.Context, functionName string, parameters any, getLog bool) *lambda.InvokeOutput {
	logType := types.LogTypeNone
	if getLog {
		logType = types.LogTypeTail
	}
	payload, err := json.Marshal(parameters)
	if err != nil {
		log.Panicf("Couldn't marshal parameters to JSON. Here's why %v\n", err)
	}
	invokeOutput, err := wrapper.LambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		LogType:      logType,
		Payload:      payload,
	})
	if err != nil {
		log.Panicf("Couldn't invoke function %v. Here's why: %v\n", functionName, err)
	}
	return invokeOutput
}

type InvokeParams struct {
	Region       *string
	Function     *string
	BaseEndpoint *string
}

func NewFunctionWrapper(params InvokeParams) FunctionWrapper {
	lambdaClient := lambda.NewFromConfig(aws.Config{
		Region:       *params.Region,
		BaseEndpoint: params.BaseEndpoint,
	})
	return FunctionWrapper{LambdaClient: lambdaClient, function: *params.Function}
}

func Invoke(r *http.Request, wrapper FunctionWrapper) *lambda.InvokeOutput {

	out := map[string]any{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Panicf("Couldn't read request body. Here's why: %v\n", err)
	}

	out["body"] = string(body)
	out["httpMethod"] = r.Method
	out["path"] = r.URL.Path
	// headerStr := json.M(r.Header)
	head := map[string]string{}
	for k, v := range r.Header {
		// fmt.Printf("key[%s] value[%s]\n", k, v)
		head[k] = v[0]
	}
	out["headers"] = head
	pp := map[string]string{}
	for k, v := range r.URL.Query() {
		// fmt.Printf("key[%s] value[%s]\n", k, v)
		pp[k] = v[0]
	}
	out["queryStringParameters"] = pp

	// Invoke the Lambda function
	invokeOutput := wrapper.Invoke(context.TODO(), wrapper.function, out, false)
	// log.Printf("Invocation result: %v\n", string(invokeOutput.Payload))
	return invokeOutput
}
