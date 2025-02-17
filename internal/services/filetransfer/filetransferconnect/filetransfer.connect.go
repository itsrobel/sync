// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: filetransfer/filetransfer.proto

package filetransferconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	filetransfer "github.com/itsrobel/sync/internal/services/filetransfer"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// FileServiceName is the fully-qualified name of the FileService service.
	FileServiceName = "filetransfer.FileService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// FileServiceControlStreamProcedure is the fully-qualified name of the FileService's ControlStream
	// RPC.
	FileServiceControlStreamProcedure = "/filetransfer.FileService/ControlStream"
	// FileServiceSendFileToServerProcedure is the fully-qualified name of the FileService's
	// SendFileToServer RPC.
	FileServiceSendFileToServerProcedure = "/filetransfer.FileService/SendFileToServer"
	// FileServiceGreetProcedure is the fully-qualified name of the FileService's Greet RPC.
	FileServiceGreetProcedure = "/filetransfer.FileService/Greet"
	// FileServiceRetrieveListOfFilesProcedure is the fully-qualified name of the FileService's
	// RetrieveListOfFiles RPC.
	FileServiceRetrieveListOfFilesProcedure = "/filetransfer.FileService/RetrieveListOfFiles"
)

// FileServiceClient is a client for the filetransfer.FileService service.
type FileServiceClient interface {
	ControlStream(context.Context) *connect.BidiStreamForClient[filetransfer.ControlMessage, filetransfer.ControlMessage]
	SendFileToServer(context.Context) *connect.ClientStreamForClient[filetransfer.FileVersionData, filetransfer.ActionResponse]
	Greet(context.Context, *connect.Request[filetransfer.GreetRequest]) (*connect.Response[filetransfer.GreetResponse], error)
	RetrieveListOfFiles(context.Context, *connect.Request[filetransfer.ActionRequest]) (*connect.Response[filetransfer.FileList], error)
}

// NewFileServiceClient constructs a client for the filetransfer.FileService service. By default, it
// uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewFileServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) FileServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	fileServiceMethods := filetransfer.File_filetransfer_filetransfer_proto.Services().ByName("FileService").Methods()
	return &fileServiceClient{
		controlStream: connect.NewClient[filetransfer.ControlMessage, filetransfer.ControlMessage](
			httpClient,
			baseURL+FileServiceControlStreamProcedure,
			connect.WithSchema(fileServiceMethods.ByName("ControlStream")),
			connect.WithClientOptions(opts...),
		),
		sendFileToServer: connect.NewClient[filetransfer.FileVersionData, filetransfer.ActionResponse](
			httpClient,
			baseURL+FileServiceSendFileToServerProcedure,
			connect.WithSchema(fileServiceMethods.ByName("SendFileToServer")),
			connect.WithClientOptions(opts...),
		),
		greet: connect.NewClient[filetransfer.GreetRequest, filetransfer.GreetResponse](
			httpClient,
			baseURL+FileServiceGreetProcedure,
			connect.WithSchema(fileServiceMethods.ByName("Greet")),
			connect.WithClientOptions(opts...),
		),
		retrieveListOfFiles: connect.NewClient[filetransfer.ActionRequest, filetransfer.FileList](
			httpClient,
			baseURL+FileServiceRetrieveListOfFilesProcedure,
			connect.WithSchema(fileServiceMethods.ByName("RetrieveListOfFiles")),
			connect.WithClientOptions(opts...),
		),
	}
}

// fileServiceClient implements FileServiceClient.
type fileServiceClient struct {
	controlStream       *connect.Client[filetransfer.ControlMessage, filetransfer.ControlMessage]
	sendFileToServer    *connect.Client[filetransfer.FileVersionData, filetransfer.ActionResponse]
	greet               *connect.Client[filetransfer.GreetRequest, filetransfer.GreetResponse]
	retrieveListOfFiles *connect.Client[filetransfer.ActionRequest, filetransfer.FileList]
}

// ControlStream calls filetransfer.FileService.ControlStream.
func (c *fileServiceClient) ControlStream(ctx context.Context) *connect.BidiStreamForClient[filetransfer.ControlMessage, filetransfer.ControlMessage] {
	return c.controlStream.CallBidiStream(ctx)
}

// SendFileToServer calls filetransfer.FileService.SendFileToServer.
func (c *fileServiceClient) SendFileToServer(ctx context.Context) *connect.ClientStreamForClient[filetransfer.FileVersionData, filetransfer.ActionResponse] {
	return c.sendFileToServer.CallClientStream(ctx)
}

// Greet calls filetransfer.FileService.Greet.
func (c *fileServiceClient) Greet(ctx context.Context, req *connect.Request[filetransfer.GreetRequest]) (*connect.Response[filetransfer.GreetResponse], error) {
	return c.greet.CallUnary(ctx, req)
}

// RetrieveListOfFiles calls filetransfer.FileService.RetrieveListOfFiles.
func (c *fileServiceClient) RetrieveListOfFiles(ctx context.Context, req *connect.Request[filetransfer.ActionRequest]) (*connect.Response[filetransfer.FileList], error) {
	return c.retrieveListOfFiles.CallUnary(ctx, req)
}

// FileServiceHandler is an implementation of the filetransfer.FileService service.
type FileServiceHandler interface {
	ControlStream(context.Context, *connect.BidiStream[filetransfer.ControlMessage, filetransfer.ControlMessage]) error
	SendFileToServer(context.Context, *connect.ClientStream[filetransfer.FileVersionData]) (*connect.Response[filetransfer.ActionResponse], error)
	Greet(context.Context, *connect.Request[filetransfer.GreetRequest]) (*connect.Response[filetransfer.GreetResponse], error)
	RetrieveListOfFiles(context.Context, *connect.Request[filetransfer.ActionRequest]) (*connect.Response[filetransfer.FileList], error)
}

// NewFileServiceHandler builds an HTTP handler from the service implementation. It returns the path
// on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewFileServiceHandler(svc FileServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	fileServiceMethods := filetransfer.File_filetransfer_filetransfer_proto.Services().ByName("FileService").Methods()
	fileServiceControlStreamHandler := connect.NewBidiStreamHandler(
		FileServiceControlStreamProcedure,
		svc.ControlStream,
		connect.WithSchema(fileServiceMethods.ByName("ControlStream")),
		connect.WithHandlerOptions(opts...),
	)
	fileServiceSendFileToServerHandler := connect.NewClientStreamHandler(
		FileServiceSendFileToServerProcedure,
		svc.SendFileToServer,
		connect.WithSchema(fileServiceMethods.ByName("SendFileToServer")),
		connect.WithHandlerOptions(opts...),
	)
	fileServiceGreetHandler := connect.NewUnaryHandler(
		FileServiceGreetProcedure,
		svc.Greet,
		connect.WithSchema(fileServiceMethods.ByName("Greet")),
		connect.WithHandlerOptions(opts...),
	)
	fileServiceRetrieveListOfFilesHandler := connect.NewUnaryHandler(
		FileServiceRetrieveListOfFilesProcedure,
		svc.RetrieveListOfFiles,
		connect.WithSchema(fileServiceMethods.ByName("RetrieveListOfFiles")),
		connect.WithHandlerOptions(opts...),
	)
	return "/filetransfer.FileService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case FileServiceControlStreamProcedure:
			fileServiceControlStreamHandler.ServeHTTP(w, r)
		case FileServiceSendFileToServerProcedure:
			fileServiceSendFileToServerHandler.ServeHTTP(w, r)
		case FileServiceGreetProcedure:
			fileServiceGreetHandler.ServeHTTP(w, r)
		case FileServiceRetrieveListOfFilesProcedure:
			fileServiceRetrieveListOfFilesHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedFileServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedFileServiceHandler struct{}

func (UnimplementedFileServiceHandler) ControlStream(context.Context, *connect.BidiStream[filetransfer.ControlMessage, filetransfer.ControlMessage]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("filetransfer.FileService.ControlStream is not implemented"))
}

func (UnimplementedFileServiceHandler) SendFileToServer(context.Context, *connect.ClientStream[filetransfer.FileVersionData]) (*connect.Response[filetransfer.ActionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("filetransfer.FileService.SendFileToServer is not implemented"))
}

func (UnimplementedFileServiceHandler) Greet(context.Context, *connect.Request[filetransfer.GreetRequest]) (*connect.Response[filetransfer.GreetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("filetransfer.FileService.Greet is not implemented"))
}

func (UnimplementedFileServiceHandler) RetrieveListOfFiles(context.Context, *connect.Request[filetransfer.ActionRequest]) (*connect.Response[filetransfer.FileList], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("filetransfer.FileService.RetrieveListOfFiles is not implemented"))
}
