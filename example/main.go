package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// EchoGreeterServer has implemented the GreeterServer interface that created from the service in proto file.
type EchoGreeterServer struct {
}

// SayHello implements the GreeterServer interface method.
// SayHello returns a greeting to the name sent.
func (s *EchoGreeterServer) SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error) {
	return &HelloReply{
		Message: fmt.Sprintf("Hello, %s!", req.Name),
	}, nil
}

func main() {
	// Create the GreeterServer.
	srv := &EchoGreeterServer{}

	// Create the GreeterHTTPConverter generated by protoc-gen-gohttp.
	// This converter converts the GreeterServer interface that created from the service in proto to http.HandlerFunc.
	conv := NewGreeterHTTPConverter(srv)

	// Register SayHello HandlerFunc to the server.
	// If you do not need a callback, pass nil as argument.
	http.Handle("/sayhello", conv.SayHello(logCallback))
	// If you want to create a path from Proto's service name and method name, use the SayHelloWithName method.
	// In this case, the strings 'Greeter' and 'SayHello' are returned.
	http.Handle(restPath(conv.SayHelloWithName(logCallback)))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// logCallback is called when exiting ServeHTTP
// and receives Context, ResponseWriter, Request, service argument, service return value and error.
func logCallback(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error) {
	log.Printf("INFO: call %s: arg: {%v}, ret: {%s}", r.RequestURI, arg, ret)
	// YOU MUST HANDLE ERROR
	if err != nil {
		log.Printf("ERROR: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		p := status.New(codes.Unknown, err.Error()).Proto()
		switch r.Header.Get("Content-Type") {
		case "application/protobuf", "application/x-protobuf":
			buf, err := proto.Marshal(p)
			if err != nil {
				return
			}
			if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {
				return
			}
		case "application/json":
			if err := json.NewEncoder(w).Encode(p); err != nil {
				return
			}
		default:
		}
	}
}

func restPath(service, method string, hf http.HandlerFunc) (string, http.HandlerFunc) {
	return fmt.Sprintf("/%s/%s", strings.ToLower(service), strings.ToLower(method)), hf
}
