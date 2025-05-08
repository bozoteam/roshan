package main

import (
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
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	fmt.Printf("Is development=%v\n", helpers.IsDevelopment)

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

	allowedOrigins := []string{
		"https://bozo.mateusbento.com",
	}

	if helpers.IsDevelopment {
		allowedOrigins = append(allowedOrigins, []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:50000",
			"http://127.0.0.1:50000",
			"http://bozo.mateusbento.com",
		}...)
	}

	// add cors middleware
	ginRouter.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))
	// }

	ginRouter.GET("/api/v1/chat/rooms/:id/ws", func(ctx *gin.Context) {
		chatUsecase.HandleWebSocket(ctx)
	})

	ginRouter.GET("/api/v1/health", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "OK")
	})

	ginRouter.GET("/health", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "OK")
	})

	ginRouter.NoRoute(func(ctx *gin.Context) {
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	})

	listener, err := net.Listen("tcp4", "0.0.0.0:8080")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = ginRouter.RunListener(listener)
	if err != nil {
		fmt.Println("Error starting server:", err)
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
