package db

import (
	"context"
	"log"
	"math/rand"

	"github.com/chisty/gopherhub/internal/store"
)

var usernames = []string{
	"alex_smith", "emma_wilson", "james_brown", "sophia_lee", "lucas_martin",
	"olivia_jones", "noah_taylor", "mia_anderson", "ethan_white", "ava_thomas",
	"liam_garcia", "isabella_moore", "mason_clark", "zoe_walker", "owen_hill",
	"lily_scott", "jack_green", "chloe_adams", "ryan_baker", "luna_morris",
}

var postTitles = []string{
	"Benefits of Morning Walks", "Easy Pasta Recipes", "Home Office Setup", "Pet Care Basics", "Gardening Tips",
	"Travel on a Budget", "Healthy Sleep Habits", "Basic Photography", "Quick Workout Guide", "Home Organization",
	"Simple Meditation", "Book Reading Benefits", "Coffee Brewing Guide", "Indoor Plant Care", "Time Management",
	"Writing Better Emails", "Clean Your Space", "Mindful Living Tips", "Music for Focus", "Digital Wellness",
}

var postContents = []string{
	"Walking 30 minutes daily improves health.", "Simple pasta recipe with tomatoes and herbs.", "Ergonomic desk setup guide for WFH.", "Essential tips for new pet owners.", "Basic tips for growing vegetables.",
	"Save money while exploring new places.", "How to improve your sleep quality.", "Learn basic camera settings.", "10-minute effective exercises.", "Keep your home clean and organized.",
	"Start meditation with 5 minutes daily.", "Reading improves mental sharpness.", "Perfect coffee brewing steps.", "Best low-maintenance indoor plants.", "Manage your time effectively.",
	"Write clear and professional emails.", "Simple steps to reduce clutter.", "Practice mindfulness daily.", "Playlists to boost productivity.", "Reduce screen time and stay focused.",
}

func Seed(store store.Storage) error {
	ctx := context.Background()

	users := generateUsers(20)
	for _, user := range users {
		if err := store.Users.Create(ctx, user); err != nil {
			log.Println("Error creating user", err)
			return err
		}
	}

	log.Println("Users created")

	posts := generatePosts(100, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Println("Error creating post", err)
			return err
		}
	}

	log.Println("Posts created")

	comments := generateComments(300, users, posts)
	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			log.Println("Error creating comment", err)
			return err
		}
	}

	log.Println("Comments added")

	return nil
}

func generateUsers(n int) []*store.User {
	users := make([]*store.User, n)
	for i := 0; i < n; i++ {
		users[i] = &store.User{
			Username: usernames[i],
			Email:    usernames[i] + "@example.com",
			Password: "password",
		}
	}
	return users
}

func generatePosts(n int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, n)
	for i := 0; i < n; i++ {
		user := users[rand.Intn(len(users))]
		posts[i] = &store.Post{
			Title:   postTitles[rand.Intn(len(postTitles))],
			Content: postContents[rand.Intn(len(postContents))],
			UserID:  user.ID,
			Tags: []string{
				postTitles[rand.Intn(len(postTitles))],
				postTitles[rand.Intn(len(postTitles))],
			},
		}
	}
	return posts
}

func generateComments(n int, users []*store.User, posts []*store.Post) []*store.Comment {
	comments := make([]*store.Comment, n)
	for i := 0; i < n; i++ {
		user := users[rand.Intn(len(users))]
		post := posts[rand.Intn(len(posts))]
		comments[i] = &store.Comment{
			Content: postContents[rand.Intn(len(postContents))],
			UserID:  user.ID,
			PostID:  post.ID,
		}
	}
	return comments
}
