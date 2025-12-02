package atdb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetOneDoc mengambil satu dokumen dari collection
func GetOneDoc[T any](db *mongo.Database, collection string, filter bson.M) (T, error) {
	var result T
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := db.Collection(collection).FindOne(ctx, filter).Decode(&result)
	return result, err
}

// GetAllDoc mengambil semua dokumen dari collection dengan filter
func GetAllDoc[T any](db *mongo.Database, collection string, filter bson.M) ([]T, error) {
	var results []T
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := db.Collection(collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var elem T
		if err := cursor.Decode(&elem); err != nil {
			continue
		}
		results = append(results, elem)
	}

	return results, cursor.Err()
}

// GetAllDocWithSort mengambil semua dokumen dengan sorting
func GetAllDocWithSort[T any](db *mongo.Database, collection string, filter bson.M, sort bson.D) ([]T, error) {
	var results []T
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := options.Find().SetSort(sort)
	cursor, err := db.Collection(collection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var elem T
		if err := cursor.Decode(&elem); err != nil {
			continue
		}
		results = append(results, elem)
	}

	return results, cursor.Err()
}

// InsertOneDoc menyisipkan satu dokumen ke collection
func InsertOneDoc(db *mongo.Database, collection string, doc interface{}) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := db.Collection(collection).InsertOne(ctx, doc)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return result.InsertedID.(primitive.ObjectID), nil
}

// UpdateOneDoc mengupdate satu dokumen
func UpdateOneDoc(db *mongo.Database, collection string, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return db.Collection(collection).UpdateOne(ctx, filter, bson.M{"$set": update})
}

// ReplaceOneDoc mengganti satu dokumen
func ReplaceOneDoc(db *mongo.Database, collection string, filter bson.M, replacement interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return db.Collection(collection).ReplaceOne(ctx, filter, replacement)
}

// DeleteOneDoc menghapus satu dokumen
func DeleteOneDoc(db *mongo.Database, collection string, filter bson.M) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return db.Collection(collection).DeleteOne(ctx, filter)
}

// CountDoc menghitung dokumen dalam collection
func CountDoc(db *mongo.Database, collection string, filter bson.M) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return db.Collection(collection).CountDocuments(ctx, filter)
}