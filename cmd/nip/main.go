package main

import (
	"context"
	"database/sql"
	"log"
	"nip"
	"nip/internal/repository/mongodb"
	"nip/internal/repository/mysql"
	"nip/internal/service"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Critical system failure: %v", err)
	}
}

func run() error {
	log.Println("Initializing Nip API Services...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sqlDB := initializeMysql()
	defer sqlDB.Close()

	mongoClient, mongoDB := initializeMongoDB(ctx)
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}()

	// Initialize Repositories (internal implementations)
	userRepo := mysql.NewUserRepo(sqlDB)
	tweetRepo := mongodb.NewTweetRepo(mongoDB)
	followRepo := mysql.NewFollowRepository(sqlDB)

	// Initialize Services (Dependency Injection from internal/service)
	otpSvc := service.NewOtpService()
	emailSvc := service.NewEmailService()
	phoneSvc := service.NewPhoneService()
	notificationSvc := service.NewNotificationService(userRepo, emailSvc, phoneSvc)

	// Wire Up Core Business Services
	authService := service.NewAuthService(userRepo, otpSvc, emailSvc, phoneSvc)
	timelineService := service.NewTimelineService(ctx, tweetRepo, userRepo, followRepo)
	tweetService := service.NewTweetService(tweetRepo, userRepo, timelineService, notificationSvc)
	_ = service.NewUserService(userRepo, timelineService, followRepo)

	log.Println("Nip Services Ready.")
	_ = authService
	_ = tweetService

	// Keep service running or handle idle state
	// For now, we exit cleanly as this is a backend service shell
	return nil
}

func initializeMysql() *sql.DB {
	mysqlURI := "root:root@tcp(127.0.0.1:3306)/nip?parseTime=true"
	sqlDB, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatalf("Failed to open MySQL connection: %v", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		log.Printf("Warning: MySQL connection check failed: %v", err)
	} else {
		log.Println("Successfully connected to MySQL database")
	}

	return sqlDB
}

func initializeMongoDB(ctx context.Context) (*mongo.Client, *mongo.Database) {
	mongoURI := "mongodb://127.0.0.1:27017"

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Printf("Warning: MongoDB connection check failed: %v", err)
	} else {
		log.Println("Successfully connected to MongoDB database")
	}

	return mongoClient, mongoClient.Database("nip")
}

func signUpCycle(ctx context.Context,
	authService nip.AuthService, userName string, emailOrPhoneNumber string) (uint, error) {

	_, err := authService.SignUp(ctx, userName, emailOrPhoneNumber)
	if err != nil {
		return 0, err
	}

	otpCode, err := authService.SendOtp(ctx, emailOrPhoneNumber)
	if err != nil {
		return 0, err
	}

	user, err := authService.Login(ctx, emailOrPhoneNumber, otpCode)
	if err != nil {
		return 0, err
	}

	return user.Id, nil
}
