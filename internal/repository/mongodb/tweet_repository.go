package mongodb

import (
	"context"
	"errors"
	"fmt"
	"nip"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type tweetRepository struct {
	col *mongo.Collection
}

type tweetDoc struct {
	Id               primitive.ObjectID `bson:"_id,omitempty"`
	UserId           uint               `bson:"userId"`
	Content          string             `bson:"content"`
	Hashtags         []string           `bson:"hashtags"`
	MentionedUserIds []uint             `bson:"mentionedUserIds"`
	CreatedAt        time.Time          `bson:"createdAt"`
}

// NewTweetRepo creates a new MongoDB implementation of nip.TweetRepository.
func NewTweetRepo(db *mongo.Database) nip.TweetRepository {
	return &tweetRepository{col: db.Collection("tweets")}
}

// Create inserts a new tweet document into MongoDB.
func (r *tweetRepository) Create(ctx context.Context, t nip.Tweet) (string, error) {
	if t.Content == "" {
		return "", errors.New("content is required")
	}

	res, err := r.col.InsertOne(ctx, tweetDoc{
		UserId:           t.UserId,
		Content:          t.Content,
		Hashtags:         t.Hashtags,
		MentionedUserIds: t.MentionedUserIds,
		CreatedAt:        time.Now().UTC(),
	})
	if err != nil {
		return "", fmt.Errorf("Error in creating tweet: %w", err)
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("unexpected inserted id type")
	}
	return oid.Hex(), nil
}

// ListAllLatest retrieves all tweets sorted by creation time descending.
func (r *tweetRepository) ListAllLatest(ctx context.Context) ([]nip.Tweet, error) {
	cur, err := r.col.Find(ctx, bson.M{}, options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	out, err := r.list(ctx, cur)
	cur.Close(ctx)

	return out, err
}

// ListLatest retrieves tweets for a specific set of IDs, sorted by creation time descending.
func (r *tweetRepository) ListLatest(ctx context.Context, tweetIds []string) ([]nip.Tweet, error) {
	var objectIds []primitive.ObjectID
	for _, id := range tweetIds {
		objId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objectIds = append(objectIds, objId)
	}

	if len(objectIds) == 0 {
		return []nip.Tweet{}, nil
	}

	filter := bson.M{"_id": bson.M{"$in": objectIds}}

	cur, err := r.col.Find(ctx, filter, options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	out, err := r.list(ctx, cur)
	cur.Close(ctx)

	return out, err
}

// ListIDsByUserID retrieves the ObjectID strings for all tweets belonging to a user.
func (r *tweetRepository) ListIDsByUserID(ctx context.Context, userId uint) ([]string, error) {
	filter := bson.M{"userId": userId}
	opts := options.Find().SetProjection(bson.M{"_id": 1})

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find tweets for user %d: %w", userId, err)
	}
	defer cur.Close(ctx)

	var out []string
	for cur.Next(ctx) {
		var d struct {
			Id primitive.ObjectID `bson:"_id"`
		}
		if err := cur.Decode(&d); err != nil {
			return nil, err
		}
		out = append(out, d.Id.Hex())
	}
	return out, cur.Err()
}

// GetByID finds a tweet by its ObjectID string. Returns nip.ErrNotFound if missing.
func (r *tweetRepository) GetByID(ctx context.Context, tweetId string) (nip.Tweet, error) {
	objectId, err := primitive.ObjectIDFromHex(tweetId)
	if err != nil {
		return nip.Tweet{}, fmt.Errorf("invalid tweet id format: %w", err)
	}

	var d tweetDoc
	err = r.col.FindOne(ctx, bson.M{"_id": objectId}).Decode(&d)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nip.Tweet{}, nip.ErrNotFound
		}
		return nip.Tweet{}, err
	}

	return nip.Tweet{
		Id:               d.Id.Hex(),
		UserId:           d.UserId,
		Content:          d.Content,
		Hashtags:         d.Hashtags,
		MentionedUserIds: d.MentionedUserIds,
		CreatedAt:        d.CreatedAt,
	}, nil
}

func (r *tweetRepository) list(ctx context.Context, cur *mongo.Cursor) ([]nip.Tweet, error) {
	var out []nip.Tweet
	for cur.Next(ctx) {
		var d tweetDoc
		if err := cur.Decode(&d); err != nil {
			return nil, err
		}
		out = append(out, nip.Tweet{
			Id:               d.Id.Hex(),
			UserId:           d.UserId,
			Content:          d.Content,
			Hashtags:         d.Hashtags,
			MentionedUserIds: d.MentionedUserIds,
			CreatedAt:        d.CreatedAt,
		})
	}
	return out, cur.Err()
}
