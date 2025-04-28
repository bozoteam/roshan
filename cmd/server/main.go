package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"

	"google.golang.org/grpc/encoding"

	"connectrpc.com/vanguard"
	"connectrpc.com/vanguard/vanguardgrpc"
	database "github.com/bozoteam/roshan/adapter/database"
	authGen "github.com/bozoteam/roshan/adapter/grpc/gen/auth"
	chatGen "github.com/bozoteam/roshan/adapter/grpc/gen/chat"
	userGen "github.com/bozoteam/roshan/adapter/grpc/gen/user"
	auth_service "github.com/bozoteam/roshan/adapter/grpc/service/auth"
	chat_service "github.com/bozoteam/roshan/adapter/grpc/service/chat"
	user_service "github.com/bozoteam/roshan/adapter/grpc/service/user"
	"github.com/bozoteam/roshan/helpers"
	"github.com/bozoteam/roshan/modules/auth/middlewares"
	jwtRepository "github.com/bozoteam/roshan/modules/auth/repository/jwt"
	authUsecase "github.com/bozoteam/roshan/modules/auth/usecase"
	chatUsecase "github.com/bozoteam/roshan/modules/chat/usecase"
	userRepository "github.com/bozoteam/roshan/modules/user/repository"
	userUsecase "github.com/bozoteam/roshan/modules/user/usecase"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// RunServer starts the API server
func RunServer() {
	blacklistedPaths, err := protoOptionToShouldPermission(authGen.File_auth_auth_proto, userGen.File_user_user_proto)
	if err != nil {
		panic(err)
	}
	fmt.Println(blacklistedPaths)

	db := database.GetDBConnection()
	userRepository := userRepository.NewUserRepository(db)
	jwtRepository := jwtRepository.NewJWTRepository()
	authUsecase := authUsecase.NewAuthUsecase(userRepository, jwtRepository)
	authMiddleware := middlewares.NewAuthMiddleware(jwtRepository, userRepository, blacklistedPaths)
	chatUsecase := chatUsecase.NewChatUsecase(userRepository, jwtRepository)
	userUsecase := userUsecase.NewUserUsecase(db)

	authInterceptor := authMiddleware.UnaryInterceptor

	encoding.RegisterCodec(vanguardgrpc.NewCodec(&vanguard.JSONCodec{
		MarshalOptions:   protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true},
		UnmarshalOptions: protojson.UnmarshalOptions{DiscardUnknown: true},
	}))

	server := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor),
	)
	// Create the Connect service implementation
	authService := auth_service.NewAuthService(authUsecase)
	userService := user_service.NewUserService(userUsecase)
	chatService := chat_service.NewChatService(chatUsecase)

	authGen.RegisterAuthServiceServer(server, authService)
	userGen.RegisterUserServiceServer(server, userService)
	chatGen.RegisterChatServiceServer(server, chatService)

	handler, err := vanguardgrpc.NewTranscoder(server)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ginRouter := gin.Default()

	// Combine handlers
	combinedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", helpers.GetEnv("CORS_ALLOWED_ORIGINS"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if strings.HasSuffix(r.URL.Path, "/ws") {
			ginRouter.ServeHTTP(w, r)
			return
		}
		handler.ServeHTTP(w, r)
	})

	ginRouter.GET("/api/v1/chat/rooms/:id/ws", func(ctx *gin.Context) {
		chatUsecase.HandleWebSocket(ctx)
	})

	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = http.Serve(listener, h2c.NewHandler(combinedHandler, &http2.Server{}))
	if !errors.Is(err, http.ErrServerClosed) {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	helpers.LoadDotEnv()

	RunServer()
}

func protoOptionToShouldPermission(fd ...protoreflect.FileDescriptor) (map[string]struct{}, error) {
	var (
		optionName = "required"
	)

	blacklistedPaths := map[string]struct{}{}

	for _, f := range fd {
		services := f.Services()
		for y := range services.Len() {
			for x := range services.Get(y).Methods().Len() {
				method := services.Get(y).Methods().Get(x)
				methodName := string(method.FullName())
				methodNameIdx := strings.LastIndex(string(method.FullName()), ".")
				methodName = "/" + methodName[:methodNameIdx] + "/" + methodName[methodNameIdx+1:]

				opts, ok := method.Options().(*descriptorpb.MethodOptions)
				if ok {
					proto.RangeExtensions(opts, func(et protoreflect.ExtensionType, i any) bool {
						if (et.TypeDescriptor().Name()) == protoreflect.Name(optionName) {
							value := reflect.ValueOf(i)
							if value.Kind() == reflect.Bool && !value.Bool() {
								blacklistedPaths[methodName] = struct{}{}
							}
						}
						return true
					})
				}
			}
		}
	}

	return blacklistedPaths, nil
}
