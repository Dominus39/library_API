package utils

import (
	"context"
	"fmt"
	"gc2-yugo/entity"
	"time"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func StartSchedulerJob() {
	// Create a new cron job scheduler
	c := cron.New()

	// Schedule the job to run every day at midnight (00:00)
	c.AddFunc("0 0 * * *", checkAndUpdateLateBooks)

	// Start the cron scheduler
	c.Start()

	// Keep the main function running to allow the cron job to continue
	select {}
}

func checkAndUpdateLateBooks() {
	// Create a MongoDB client and context
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		fmt.Println("Error connecting to MongoDB:", err)
		return
	}
	defer client.Disconnect(context.Background())

	// Access the books collection
	booksCollection := client.Database("GC2").Collection("books")

	// Access the borrowedBooks collection to get the list of borrowed books
	borrowedBooksCollection := client.Database("yourDatabaseName").Collection("borrowed_books")

	// Get the current date
	now := time.Now()

	// Find all borrowed books
	cursor, err := borrowedBooksCollection.Find(context.Background(), bson.M{
		"return_date": bson.M{"$lt": now.Format(time.RFC3339)}, // Find books where return_date is less than the current date
		"status":      "borrowed",
	})
	if err != nil {
		fmt.Println("Error finding borrowed books:", err)
		return
	}
	defer cursor.Close(context.Background())

	// Iterate through the cursor and check each borrowed book
	for cursor.Next(context.Background()) {
		var borrowedBook entity.BorrowedBooks
		err := cursor.Decode(&borrowedBook)
		if err != nil {
			fmt.Println("Error decoding borrowed book:", err)
			continue
		}

		// Check if the book status is still "borrowed"
		bookCursor, err := booksCollection.FindOne(context.Background(), bson.M{"_id": borrowedBook.BookID}).DecodeBytes()
		if err != nil {
			fmt.Println("Error finding book:", err)
			continue
		}

		// If the book's status is "borrowed", update it to "Late"
		if bookCursor != nil {
			_, err := booksCollection.UpdateOne(
				context.Background(),
				bson.M{"_id": borrowedBook.BookID},
				bson.M{"$set": bson.M{"status": "Late"}},
			)
			if err != nil {
				fmt.Println("Error updating book status:", err)
			} else {
				fmt.Println("Book status updated to Late:", borrowedBook.BookID)
			}
		}
	}

	// Check for errors during iteration
	if err := cursor.Err(); err != nil {
		fmt.Println("Error iterating over cursor:", err)
	}
}
