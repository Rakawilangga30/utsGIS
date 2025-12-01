package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/* =======================================================
   GLOBAL MONGO + GRIDFS + SESSION
======================================================= */

var mongoDB *mongo.Database
var photoBucket *gridfs.Bucket
var Store *sessions.CookieStore

var Users *mongo.Collection
var Places *mongo.Collection
var Reviews *mongo.Collection

/* =======================================================
   INIT (ENV + MONGO)
======================================================= */

func init() {
	godotenv.Load()

	uri := os.Getenv("MONGO_URI")
	db := os.Getenv("DB_NAME")
	secret := os.Getenv("SESSION_KEY")

	if secret == "" {
		secret = "MYTRAVEL_SECRET"
	}

	Store = sessions.NewCookieStore([]byte(secret))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	mongoDB = client.Database(db)
	photoBucket, _ = gridfs.NewBucket(mongoDB)

	Users = mongoDB.Collection("Users")
	Places = mongoDB.Collection("Data_Travel")
	Reviews = mongoDB.Collection("Reviews")

	fmt.Println("MongoDB Connected:", db)
}

/* =======================================================
   HELPERS
======================================================= */

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func EnableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
}

func getSession(r *http.Request) (*sessions.Session, error) {
	return Store.Get(r, "mytravel-session")
}

func requireLogin(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	sess, _ := getSession(r)
	id, ok1 := sess.Values["user_id"].(string)
	role, ok2 := sess.Values["role"].(string)

	if !ok1 || !ok2 {
		jsonResponse(w, 401, map[string]string{"error": "unauthorized"})
		return "", "", false
	}

	return id, role, true
}

/* =======================================================
   MAIN ROUTER (ENTRY FOR VERCEL)
======================================================= */

func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		EnableCORS(w)
		if r.Method == "OPTIONS" {
			return
		}

		p := r.URL.Path

		/* -------- AUTH -------- */
		if p == "/api/auth/register" {
			Register(w, r); return
		}
		if p == "/api/auth/login" {
			Login(w, r); return
		}
		if p == "/api/auth/logout" {
			Logout(w, r); return
		}
		if p == "/api/auth/me" {
			Me(w, r); return
		}

		/* -------- PLACES -------- */
		if p == "/api/places" && r.Method == "GET" {
			GetPlaces(w, r); return
		}
		if p == "/api/places" && r.Method == "POST" {
			CreatePlace(w, r); return
		}
		if strings.HasPrefix(p, "/api/places/") && r.Method == "PUT" {
			UpdatePlace(w, r); return
		}
		if strings.HasPrefix(p, "/api/places/") && r.Method == "DELETE" {
			DeletePlace(w, r); return
		}

		/* -------- REVIEWS -------- */
		if p == "/api/reviews" && r.Method == "GET" {
			GetReviews(w, r); return
		}
		if p == "/api/reviews" && r.Method == "POST" {
			CreateReview(w, r); return
		}
		if strings.HasPrefix(p, "/api/reviews/") && r.Method == "DELETE" {
			DeleteReview(w, r); return
		}

		/* -------- PHOTO -------- */
		if strings.HasPrefix(p, "/api/photo/") {
			GetPhoto(w, r); return
		}

		jsonResponse(w, 404, map[string]string{"error": "route not found"})
	}
}

/* =======================================================
   AUTH: REGISTER + LOGIN + LOGOUT + ME
======================================================= */

func Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&body)

	if body.Email == "" || body.Password == "" {
		jsonResponse(w, 400, map[string]string{"error": "email & password required"})
		return
	}

	count, _ := Users.CountDocuments(context.Background(), bson.M{"email": body.Email})
	if count > 0 {
		jsonResponse(w, 400, map[string]string{"error": "email already used"})
		return
	}

	Users.InsertOne(context.Background(), bson.M{
		"name":       body.Name,
		"email":      body.Email,
		"password":   body.Password,
		"role":       "user",
		"created_at": time.Now().Unix(),
	})

	jsonResponse(w, 200, map[string]string{"message": "register success"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&body)

	var user bson.M
	err := Users.FindOne(context.Background(), bson.M{"email": body.Email}).Decode(&user)
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": "email not found"})
		return
	}

	if body.Password != user["password"].(string) {
		jsonResponse(w, 400, map[string]string{"error": "wrong password"})
		return
	}

	sess, _ := getSession(r)
	sess.Values["user_id"] = user["_id"].(primitive.ObjectID).Hex()
	sess.Values["role"] = user["role"].(string)
	sess.Save(r, w)

	jsonResponse(w, 200, map[string]interface{}{
		"message": "login success",
		"name":    user["name"],
		"role":    user["role"],
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := getSession(r)
	sess.Options.MaxAge = -1
	sess.Save(r, w)

	jsonResponse(w, 200, map[string]string{"message": "logout success"})
}

func Me(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := requireLogin(w, r)
	if !ok {
		return
	}

	jsonResponse(w, 200, map[string]string{
		"user_id": uid,
		"role":    role,
	})
}

/* =======================================================
   PLACES (CRUD + UPLOAD FOTO GRIDFS)
======================================================= */

func CreatePlace(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := requireLogin(w, r)
	if !ok {
		return
	}

	r.ParseMultipartForm(10 << 20)

	name := r.FormValue("name")
	category := r.FormValue("category")
	description := r.FormValue("description")
	address := r.FormValue("address")
	latStr := r.FormValue("lat")
	lngStr := r.FormValue("lng")

	if name == "" || latStr == "" || lngStr == "" {
		jsonResponse(w, 400, map[string]string{"error": "invalid form"})
		return
	}

	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)

	file, header, err := r.FormFile("photo")
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": "photo required"})
		return
	}
	defer file.Close()

	uploadStream, err := photoBucket.OpenUploadStream(header.Filename)
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": "upload failed"})
		return
	}
	io.Copy(uploadStream, file)
	uploadStream.Close()

	photoID := uploadStream.FileID.(primitive.ObjectID).Hex()

	Places.InsertOne(context.Background(), bson.M{
		"name":        name,
		"category":    category,
		"description": description,
		"address":     address,
		"lat":         lat,
		"lng":         lng,
		"photo_id":    photoID,
		"created_by":  uid,
		"created_at":  time.Now().Unix(),
	})

	jsonResponse(w, 200, map[string]string{"message": "place created"})
}

func GetPlaces(w http.ResponseWriter, r *http.Request) {
	cursor, _ := Places.Find(context.Background(), bson.M{})
	var places []bson.M
	cursor.All(context.Background(), &places)

	jsonResponse(w, 200, places)
}

func UpdatePlace(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := requireLogin(w, r)
	if !ok {
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/places/")
	objID, _ := primitive.ObjectIDFromHex(id)

	var old bson.M
	err := Places.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&old)
	if err != nil {
		jsonResponse(w, 404, map[string]string{"error": "place not found"})
		return
	}

	// hanya admin atau pemilik yang boleh update
	if role != "admin" && old["created_by"] != uid {
		jsonResponse(w, 403, map[string]string{"error": "permission denied"})
		return
	}

	r.ParseMultipartForm(10 << 20)

	data := bson.M{
		"name":        r.FormValue("name"),
		"category":    r.FormValue("category"),
		"description": r.FormValue("description"),
		"address":     r.FormValue("address"),
	}

	lat, _ := strconv.ParseFloat(r.FormValue("lat"), 64)
	lng, _ := strconv.ParseFloat(r.FormValue("lng"), 64)

	data["lat"] = lat
	data["lng"] = lng

	// cek apakah ada foto baru
	file, header, err := r.FormFile("photo")
	if err == nil { // ada foto → upload baru
		defer file.Close()

		upload, _ := photoBucket.OpenUploadStream(header.Filename)
		io.Copy(upload, file)
		upload.Close()

		newID := upload.FileID.(primitive.ObjectID).Hex()
		data["photo_id"] = newID
	}

	// update database
	Places.UpdateOne(context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": data},
	)

	jsonResponse(w, 200, map[string]string{"message": "place updated"})
}


func DeletePlace(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := requireLogin(w, r)
	if !ok {
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/places/")
	objID, _ := primitive.ObjectIDFromHex(id)

	var place bson.M
	err := Places.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&place)
	if err != nil {
		jsonResponse(w, 404, map[string]string{"error": "place not found"})
		return
	}

	if role != "admin" && place["created_by"] != uid {
		jsonResponse(w, 403, map[string]string{"error": "permission denied"})
		return
	}

	Places.DeleteOne(context.Background(), bson.M{"_id": objID})

	jsonResponse(w, 200, map[string]string{"message": "place deleted"})
}

/* =======================================================
   REVIEWS (CRUD)
======================================================= */

func CreateReview(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := requireLogin(w, r)
	if !ok {
		return
	}

	var body struct {
		PlaceID string `json:"place_id"`
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}

	json.NewDecoder(r.Body).Decode(&body)

	Reviews.InsertOne(context.Background(), bson.M{
		"place_id":   body.PlaceID,
		"user_id":    uid,
		"rating":     body.Rating,
		"comment":    body.Comment,
		"created_at": time.Now().Unix(),
	})

	jsonResponse(w, 200, map[string]string{"message": "review added"})
}

func GetReviews(w http.ResponseWriter, r *http.Request) {
	placeID := r.URL.Query().Get("place_id")

	cursor, _ := Reviews.Find(context.Background(), bson.M{"place_id": placeID})
	var list []bson.M
	cursor.All(context.Background(), &list)

	jsonResponse(w, 200, list)
}

func DeleteReview(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := requireLogin(w, r)
	if !ok {
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/reviews/")
	objID, _ := primitive.ObjectIDFromHex(id)

	var rev bson.M
	err := Reviews.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&rev)
	if err != nil {
		jsonResponse(w, 404, map[string]string{"error": "review not found"})
		return
	}

	if role != "admin" && rev["user_id"] != uid {
		jsonResponse(w, 403, map[string]string{"error": "permission denied"})
		return
	}

	Reviews.DeleteOne(context.Background(), bson.M{"_id": objID})

	jsonResponse(w, 200, map[string]string{"message": "review deleted"})
}

/* =======================================================
   PHOTO (GET FROM GRIDFS)
======================================================= */

func GetPhoto(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/photo/")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid photo id"})
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")

	stream, err := photoBucket.OpenDownloadStream(objID)
	if err != nil {
		jsonResponse(w, 404, map[string]string{"error": "photo not found"})
		return
	}
	defer stream.Close()

	io.Copy(w, stream)
}
