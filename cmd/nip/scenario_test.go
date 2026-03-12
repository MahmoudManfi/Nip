package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"nip/internal/repository/mongodb"
	"nip/internal/repository/mysql"
	"nip/internal/service"
)

func TestNipScenario(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize Database Connections
	sqlDB := initializeMysql()
	defer sqlDB.Close()

	mongoClient, mongoDB := initializeMongoDB(ctx)
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			t.Errorf("Error disconnecting MongoDB: %v", err)
		}
	}()

	// Initialize Repositories
	userRepo := mysql.NewUserRepo(sqlDB)
	tweetRepo := mongodb.NewTweetRepo(mongoDB)
	followRepo := mysql.NewFollowRepository(sqlDB)

	// Initialize Services
	otpSvc := service.NewOtpService()
	emailSvc := service.NewEmailService()
	phoneSvc := service.NewPhoneService()
	notificationSvc := service.NewNotificationService(userRepo, emailSvc, phoneSvc)

	// Wire Up Core Business Services
	authService := service.NewAuthService(userRepo, otpSvc, emailSvc, phoneSvc)
	timelineService := service.NewTimelineService(ctx, tweetRepo, userRepo, followRepo)
	tweetService := service.NewTweetService(tweetRepo, userRepo, timelineService, notificationSvc)
	userService := service.NewUserService(userRepo, timelineService, followRepo)

	t.Run("Scenario Test", func(t *testing.T) {
		trumpName := "trump"
		trumpEmail := "trump@gmail.com"
		trumpId, err := signUpCycle(ctx, authService, trumpName, trumpEmail)
		if err != nil {
			t.Fatalf("Failed to sign up trump: %v", err)
		}

		obamaName := "obama"
		obamaPhone := "+1-111-1111"
		obamaId, err := signUpCycle(ctx, authService, obamaName, obamaPhone)
		if err != nil {
			t.Fatalf("Failed to sign up obama: %v", err)
		}

		hillaryName := "hillary"
		hillaryEmail := "hillary@hillary-dyndns.net"
		hillaryId, err := signUpCycle(ctx, authService, hillaryName, hillaryEmail)
		if err != nil {
			t.Fatalf("Failed to sign up hillary: %v", err)
		}

		if err := userService.Follow(ctx, obamaId, hillaryId); err != nil {
			t.Fatalf("Obama failed to follow Hillary: %v", err)
		}

		if err := userService.Follow(ctx, obamaId, trumpId); err != nil {
			t.Fatalf("Obama failed to follow Trump: %v", err)
		}

		if err := userService.Follow(ctx, hillaryId, trumpId); err != nil {
			t.Fatalf("Hillary failed to follow Trump: %v", err)
		}

		if _, err = tweetService.CreateTweet(ctx, trumpId, "I like this Nip thing!"); err != nil {
			t.Errorf("Trump failed to tweet: %v", err)
		}

		if _, err = tweetService.CreateTweet(ctx, obamaId, "I agree"); err != nil {
			t.Errorf("Obama failed to tweet: %v", err)
		}

		if _, err = tweetService.CreateTweet(ctx, trumpId, "I’m a goofy goober"); err != nil {
			t.Errorf("Trump failed second tweet: %v", err)
		}

		mentionContent := fmt.Sprintf("@%s we’re all goofy goobers", trumpName)
		if _, err = tweetService.CreateTweet(ctx, obamaId, mentionContent); err != nil {
			t.Errorf("Obama failed mention tweet: %v", err)
		}

		timeline, err := timelineService.GetTimeline(ctx, trumpId)
		if err != nil {
			t.Fatalf("Failed to get Trump timeline: %v", err)
		}
		fmt.Printf("%s Timeline: %d tweets found\n", trumpName, len(timeline))
		for _, v := range timeline {
			fmt.Printf("   - [User %d]: %s\n", v.UserId, v.Content)
		}

		if err := userService.Follow(ctx, trumpId, obamaId); err != nil {
			t.Fatalf("Trump failed to follow Obama: %v", err)
		}

		if _, err = tweetService.CreateTweet(ctx, obamaId, "Nip rocks"); err != nil {
			t.Errorf("Obama failed third tweet: %v", err)
		}

		mentionResponse := fmt.Sprintf("@%s that’s 1 thing we can agree on", obamaName)
		if _, err = tweetService.CreateTweet(ctx, trumpId, mentionResponse); err != nil {
			t.Errorf("Trump failed mention response: %v", err)
		}

		if _, err = tweetService.CreateTweet(ctx, hillaryId, "#NipNip Hooray"); err != nil {
			t.Errorf("Hillary failed global tweet: %v", err)
		}

		t.Run("Verify All Timelines", func(t *testing.T) {
			users := []struct {
				Id   uint
				Name string
			}{
				{trumpId, trumpName},
				{obamaId, obamaName},
				{hillaryId, hillaryName},
			}
			for _, u := range users {
				tl, err := timelineService.GetTimeline(ctx, u.Id)
				fmt.Printf("%s Timeline: %d tweets found\n", u.Name, len(tl))
				if err != nil {
					t.Errorf("Failed to get %s timeline: %v", u.Name, err)
				}
				for _, v := range tl {
					fmt.Printf("   - [User %d]: %s\n", v.UserId, v.Content)
				}
			}
		})
	})
}
