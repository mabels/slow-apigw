package frontend

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/mabels/slowapigw/awsside"
)

func Serv() {

	listen := flag.String("listen", ":18000", "The address to listen on")
	function := flag.String("function", "function", "The name of the Lambda function to invoke")
	region := flag.String("region", "us-west-2", "The region to invoke the Lambda function in")
	endpoint := flag.String("base-endpoint", "http://localhost:8001", "The base endpoint to use for the Lambda client")

	// Parse flags
	flag.Parse()

	wrapper := awsside.NewFunctionWrapper(awsside.InvokeParams{
		Region:       region,
		Function:     function,
		BaseEndpoint: endpoint,
	})

	mutex := sync.Mutex{}
	http.HandleFunc("/uploads", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		res := awsside.Invoke(r, wrapper)
		lambdaRes := struct {
			StatusCode int               `json:"statusCode"`
			Headers    map[string]string `json:"headers"`
			Body       string            `json:"body"`
		}{}
		json.Unmarshal(res.Payload, &lambdaRes)
		fmt.Printf("req: %s -> %v\n", r.URL, lambdaRes)
		for k, v := range lambdaRes.Headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(lambdaRes.StatusCode)
		w.Write([]byte(lambdaRes.Body))
		mutex.Unlock()
	})

	log.Printf("Listening on %s\n", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
