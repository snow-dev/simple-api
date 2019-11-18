package main

import (
	blogpb "github.com/snow-dev/simple-api/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	"os"
	"os/signal"

	"context"
	"fmt"
	"log"
	"net"
)

type BlogServiceServer struct{}

func (s BlogServiceServer) CreateBlog(ctx context.Context, req *blogpb.CreateBlogReq) (*blogpb.CreateBlogRes, error) {
	// Essentially doing req.Blog to access the struct with a nil check
	blog := req.GetBlog()
	// Now we have to convert it into a BlogItem type to convert into BSON
	data := BlogItem{
		//ID:		empty so it gets omitted and MongoDB generates a unique Object ID upont insertion
		AuthorID: blog.GetAuthorId(),
		Content:  blog.GetContent(),
		Title:    blog.GetTitle(),
	}

	// Insert the data into the database, result contain the newly generated Object ID for de new document.
	result, err := blogdb.InsertOne(mongoCtx, data)
	// Check for potential errors.
	if err != nil {
		// return internal gRPC error to be handled later.
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	// Add the id to the blog, first cast the "generic type" (go doesn't have real generics yet)
	oid := result.InsertedID.(primitive.ObjectID)
	// Convert the object id to it's string counterpart.
	blog.Id = oid.Hex()
	// Return the blog in a CreateBlogRes type.
	return &blogpb.CreateBlogRes{Blog: blog}, nil
}

func (s BlogServiceServer) ReadBlog(ctx context.Context, req *blogpb.ReadBlogReq) (*blogpb.ReadBlogRes, error) {
	// convert string id (from proto) to MongoDB ObjectId
	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}
	result := blogdb.FindOne(ctx, bson.M{"_id": oid})
	// Create an empty BlogItem to write our decode result to
	data := BlogItem{}
	// decode and write to data
	if err := result.Decode(&data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not find blog with Object Id %s : %v", req.GetId(), err))
	}
	//cast to reaadBlogres type
	response := &blogpb.ReadBlogRes{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: data.AuthorID,
			Title:    data.Title,
			Content:  data.Content,
		},
	}
	return response, nil
}

func (s BlogServiceServer) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogReq) (*blogpb.UpdateBlogRes, error) {
	// Get the blog data from the request.
	blog := req.GetBlog()

	//convert the Id string to a MongoDB ObjectId
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Could not convert the supplied blog id to a MongoDB ObjectId: %v", err),
		)
	}

	// Convert the data to be updated into an unordered Bson document.
	update := bson.M{
		"author_id": blog.GetAuthorId(),
		"title":     blog.GetTitle(),
		"content":   blog.GetContent(),
	}

	// Convert the oid into an unordered bson document to search by id.
	filter := bson.M{"_id": oid}

	// Result is the BSON encoded result
	// To return the updated document instead of original we have to add options.
	result := blogdb.FindOneAndUpdate(ctx, filter, bson.M{"$set": update}, options.FindOneAndUpdate().SetReturnDocument(1))

	// Decode result and write it to 'decoded'
	decoded := BlogItem{}

	err = result.Decode(&decoded)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Could not find blog with supplied ID: %v", err),
		)
	}

	return &blogpb.UpdateBlogRes{
		Blog: &blogpb.Blog{
			Id:       decoded.ID.Hex(),
			AuthorId: decoded.AuthorID,
			Title:    decoded.Title,
			Content:  decoded.Content,
		},
	}, nil

}

func (s BlogServiceServer) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogReq) (*blogpb.DeleteBlogRes, error) {
	// get the ID (string) from the request message and convert it to an abject ID
	oid, err := primitive.ObjectIDFromHex(req.GetId())
	// Check errors
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}
	// DeleteOne returns DeleteResult which is a struct containing the amount of deleted docs (in this case only 1 always)
	// So we return a boolean instead
	_, err = blogdb.DeleteOne(ctx, bson.M{"_id": oid})
	// Check errors.
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not found  find/delete blog with id %s: %v", req.GetId(), err))
	}
	// Return response with success: true if no errors is thrown (and this document is removed)
	return &blogpb.DeleteBlogRes{
		Success: true,
	}, nil
}

func (s BlogServiceServer) ListBlogs(ctx *blogpb.ListBlogsReq, stream blogpb.BlogService_ListBlogsServer) error {
	// Initiate a blog item type to write decoded data to
	data := &BlogItem{}

	// collection.Find returns a cursor for our (empty) query.
	cursor, err := blogdb.Find(context.Background(), bson.M{})
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknow internal error: %v", err))
	}
	// An expresion with defer will be called at the end of the function.
	defer cursor.Close(context.Background())
	//  cursor.Next() returns a boolean , it false there are not more items and loop will break.
	for cursor.Next(context.Background()) {
		// Decode the data at the current pointer and write it to data.
		err := cursor.Decode(data)

		if err != nil {
			return status.Errorf(codes.Unavailable, fmt.Sprintf("Could not decode data: %v", err))
		}

		// If no error is found send blog over stream.
		stream.Send(&blogpb.ListBlogsRes{
			Blog: &blogpb.Blog{
				Id:       data.ID.Hex(),
				AuthorId: data.AuthorID,
				Title:    data.Title,
				Content:  data.Content,
			},
		})
	}

	// Check if the cursor has any errors.
	if err := cursor.Err(); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknow cursor error: %v", err))
	}

	return nil

}

var db *mongo.Client
var blogdb *mongo.Collection
var mongoCtx context.Context

type BlogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson"title"`
}

func main() {
	// Configure 'log' package to give file name and line number on eg. log.Fatal
	// just the filename & line number:
	// log.SetFlags(log.Lshortfile)
	// Or add timestamps and pipe file name and line number to it:
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Starting server on port :50051...")

	// 50051 is the default port for gRPC
	// Ideally we'd use 0.0.0.0 instead of localhost as well
	listener, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatalf("Unable to listen on port :50051: %v", err)
	}

	// slice of gRPC options
	// Here we can configure things like TLS
	opts := []grpc.ServerOption{}
	// var s *grpc.Server
	s := grpc.NewServer(opts...)
	// var srv *BlogServiceServer
	srv := &BlogServiceServer{}

	blogpb.RegisterBlogServiceServer(s, srv)

	// Initialize MongoDb client
	fmt.Println("Connecting to MongoDB...")
	mongoCtx = context.Background()
	db, err = mongo.Connect(mongoCtx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping(mongoCtx, nil)
	if err != nil {
		log.Fatalf("Could not connect to MongoDB: %v\n", err)
	} else {
		fmt.Println("Connected to Mongodb")
	}

	blogdb = db.Database("test").Collection("blog")

	// Start the server in a child routine
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	fmt.Println("Server successfully started on port :50051")

	// Bad way to stop the server
	// if err := s.Serve(listener); err != nil {
	// 	log.Fatalf("Failed to serve: %v", err)
	// }
	// Right way to stop the server using a SHUTDOWN HOOK

	// Create a channel to receive OS signals
	c := make(chan os.Signal)

	// Relay os.Interrupt to our channel (os.Interrupt = CTRL+C)
	// Ignore other incoming signals
	signal.Notify(c, os.Interrupt)

	// Block main routine until a signal is received
	// As long as user doesn't press CTRL+C a message is not passed
	// And our main routine keeps running
	// If the main routine were to shutdown so would the child routine that is Serving the server
	<-c

	// After receiving CTRL+C Properly stop the server
	fmt.Println("\nStopping the server...")
	s.Stop()
	listener.Close()
	fmt.Println("Closing MongoDB connection")
	db.Disconnect(mongoCtx)
	fmt.Println("Done.")

}
