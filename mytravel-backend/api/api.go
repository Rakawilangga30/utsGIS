package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global Variables
var (
	mongoDB     *mongo.Database
	photoBucket *gridfs.Bucket
	Store       *sessions.CookieStore
	Users       *mongo.Collection
	Places      *mongo.Collection
	Reviews     *mongo.Collection
	
	// Singleton untuk Router dan DB
	routerInstance http.Handler
	once           sync.Once
)

// --- KONEKSI DB ---
func connectDB() {
	// Load .env (abaikan error jika di Vercel)
	_ = godotenv.Load()

	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")
	secret := os.Getenv("SESSION_KEY")

	if dbName == "" { dbName = "MyTravel" }
	if secret == "" { secret = "MYTRAVEL_SECRET_KEY" }
	if uri == "" { fmt.Println("Warning: MONGO_URI kosong!") }

	// Setup Session
	Store = sessions.NewCookieStore([]byte(secret))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24, // 1 hari
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println("Mongo Connect Error:", err)
		return
	}

	mongoDB = client.Database(dbName)
	photoBucket, _ = gridfs.NewBucket(mongoDB)
	Users = mongoDB.Collection("Users")
	Places = mongoDB.Collection("Data_Travel")
	Reviews = mongoDB.Collection("Reviews")

	fmt.Println("MongoDB Connected to:", dbName)
}

// --- SETUP ROUTER (Seperti InfoCuy) ---
func SetupRouter() http.Handler {
	once.Do(func() {
		connectDB() // Konek DB hanya sekali

		mux := http.NewServeMux()

		// === DEFINISI ROUTES ===
		
		// Auth
		mux.HandleFunc("/api/auth/register", Register)
		mux.HandleFunc("/api/auth/login", Login)
		mux.HandleFunc("/api/auth/logout", Logout)
		mux.HandleFunc("/api/auth/me", Me)

		// Places
		mux.HandleFunc("/api/places", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" { GetPlaces(w, r); return }
			if r.Method == "POST" { CreatePlace(w, r); return }
		})
		
		// Places with ID (Manual routing karena http.ServeMux standar terbatas)
		mux.HandleFunc("/api/places/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "PUT" { UpdatePlace(w, r); return }
			if r.Method == "DELETE" { DeletePlace(w, r); return }
			// Handle static file serving logic if any, otherwise 404
		})

		// My Places
		mux.HandleFunc("/api/my-places", GetMyPlaces)
		mux.HandleFunc("/api/my-places/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "DELETE" { DeletePlace(w, r); return }
		})

		// Reviews
		mux.HandleFunc("/api/reviews", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" { GetReviews(w, r); return }
			if r.Method == "POST" { CreateReview(w, r); return }
		})
		mux.HandleFunc("/api/reviews/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "DELETE" { DeleteReview(w, r); return }
		})

		// My Reviews
		mux.HandleFunc("/api/my-reviews", GetMyReviews)
		mux.HandleFunc("/api/my-reviews/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "DELETE" { DeleteReview(w, r); return }
		})

		// Photo
		mux.HandleFunc("/api/photo/", GetPhoto)

		// Wrap dengan CORS Middleware
		routerInstance = CORS(mux)
	})
	return routerInstance
}

// --- ENTRY POINT VERCEL ---
// Fungsi ini yang dicari oleh Vercel (Sama seperti InfoCuy)
func Handler(w http.ResponseWriter, r *http.Request) {
	router := SetupRouter()
	router.ServeHTTP(w, r)
}

// --- MIDDLEWARE CORS ---
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set Header CORS agar Frontend bisa akses
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle Preflight Request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// --- HELPERS ---

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
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

// --- FUNGSI LOGIC (SAMA SEPERTI SEBELUMNYA) ---
// (Copy paste semua fungsi handler di bawah ini: Register, Login, GetPlaces, dll)

func Register(w http.ResponseWriter, r *http.Request) {
	var body struct { Name string `json:"name"`; Email string `json:"email"`; Password string `json:"password"` }
	json.NewDecoder(r.Body).Decode(&body)
	if body.Email == "" || body.Password == "" { jsonResponse(w, 400, map[string]string{"error": "email & password required"}); return }
	count, _ := Users.CountDocuments(context.Background(), bson.M{"email": body.Email})
	if count > 0 { jsonResponse(w, 400, map[string]string{"error": "email already used"}); return }
	Users.InsertOne(context.Background(), bson.M{"name": body.Name, "email": body.Email, "password": body.Password, "role": "user", "created_at": time.Now().Unix()})
	jsonResponse(w, 200, map[string]string{"message": "register success"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var body struct { Email string `json:"email"`; Password string `json:"password"` }
	json.NewDecoder(r.Body).Decode(&body)
	var user bson.M
	err := Users.FindOne(context.Background(), bson.M{"email": body.Email}).Decode(&user)
	if err != nil { jsonResponse(w, 400, map[string]string{"error": "email not found"}); return }
	if body.Password != user["password"].(string) { jsonResponse(w, 400, map[string]string{"error": "wrong password"}); return }
	sess, _ := getSession(r)
	sess.Values["user_id"] = user["_id"].(primitive.ObjectID).Hex()
	sess.Values["role"] = user["role"].(string)
	sess.Save(r, w)
	jsonResponse(w, 200, map[string]interface{}{"message": "login success", "name": user["name"], "role": user["role"]})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := getSession(r); sess.Options.MaxAge = -1; sess.Save(r, w)
	jsonResponse(w, 200, map[string]string{"message": "logout success"})
}

func Me(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := requireLogin(w, r); if !ok { return }
	jsonResponse(w, 200, map[string]string{"user_id": uid, "role": role})
}

func CreatePlace(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := requireLogin(w, r); if !ok { return }
	r.ParseMultipartForm(10 << 20)
	name := r.FormValue("name"); latStr := r.FormValue("lat"); lngStr := r.FormValue("lng")
	lat, _ := strconv.ParseFloat(latStr, 64); lng, _ := strconv.ParseFloat(lngStr, 64)
	file, header, err := r.FormFile("photo"); photoID := ""
	if err == nil { defer file.Close(); uploadStream, _ := photoBucket.OpenUploadStream(header.Filename); io.Copy(uploadStream, file); uploadStream.Close(); photoID = uploadStream.FileID.(primitive.ObjectID).Hex() }
	Places.InsertOne(context.Background(), bson.M{"name": name, "category": r.FormValue("category"), "description": r.FormValue("description"), "address": r.FormValue("address"), "lat": lat, "lng": lng, "photo_id": photoID, "created_by": uid, "created_at": time.Now().Unix()})
	jsonResponse(w, 200, map[string]string{"message": "place created"})
}

func GetPlaces(w http.ResponseWriter, r *http.Request) {
	cursor, _ := Places.Find(context.Background(), bson.M{}); var places []bson.M; cursor.All(context.Background(), &places)
	if places == nil { places = []bson.M{} }; jsonResponse(w, 200, places)
}

func GetMyPlaces(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := requireLogin(w, r); if !ok { return }
	cursor, _ := Places.Find(context.Background(), bson.M{"created_by": uid}); var places []bson.M; cursor.All(context.Background(), &places)
	if places == nil { places = []bson.M{} }; jsonResponse(w, 200, places)
}

func UpdatePlace(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := requireLogin(w, r); if !ok { return }
	id := strings.TrimPrefix(r.URL.Path, "/api/places/"); objID, _ := primitive.ObjectIDFromHex(id)
	var old bson.M; if err := Places.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&old); err != nil { jsonResponse(w, 404, map[string]string{"error": "place not found"}); return }
	if role != "admin" && old["created_by"] != uid { jsonResponse(w, 403, map[string]string{"error": "permission denied"}); return }
	r.ParseMultipartForm(10 << 20)
	updateData := bson.M{"name": r.FormValue("name"), "category": r.FormValue("category"), "description": r.FormValue("description"), "address": r.FormValue("address"), "lat": old["lat"], "lng": old["lng"]}
	if val := r.FormValue("lat"); val != "" { f, _ := strconv.ParseFloat(val, 64); updateData["lat"] = f }
	if val := r.FormValue("lng"); val != "" { f, _ := strconv.ParseFloat(val, 64); updateData["lng"] = f }
	file, header, err := r.FormFile("photo")
	if err == nil { defer file.Close(); upload, _ := photoBucket.OpenUploadStream(header.Filename); io.Copy(upload, file); upload.Close(); updateData["photo_id"] = upload.FileID.(primitive.ObjectID).Hex() }
	Places.UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$set": updateData})
	jsonResponse(w, 200, map[string]string{"message": "place updated"})
}

func DeletePlace(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := requireLogin(w, r); if !ok { return }
	id := strings.TrimPrefix(r.URL.Path, "/api/places/"); id = strings.TrimPrefix(id, "/api/my-places/")
	objID, _ := primitive.ObjectIDFromHex(id)
	var place bson.M; if err := Places.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&place); err != nil { jsonResponse(w, 404, map[string]string{"error": "place not found"}); return }
	if role != "admin" && place["created_by"] != uid { jsonResponse(w, 403, map[string]string{"error": "permission denied"}); return }
	Places.DeleteOne(context.Background(), bson.M{"_id": objID})
	jsonResponse(w, 200, map[string]string{"message": "place deleted"})
}

func CreateReview(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := requireLogin(w, r); if !ok { return }
	var body struct { PlaceID string `json:"place_id"`; Rating int `json:"rating"`; Comment string `json:"comment"` }
	json.NewDecoder(r.Body).Decode(&body)
	Reviews.InsertOne(context.Background(), bson.M{"place_id": body.PlaceID, "user_id": uid, "rating": body.Rating, "comment": body.Comment, "created_at": time.Now().Unix()})
	jsonResponse(w, 200, map[string]string{"message": "review added"})
}

func GetReviews(w http.ResponseWriter, r *http.Request) {
	placeID := r.URL.Query().Get("place_id"); filter := bson.M{}
	if placeID != "" { filter = bson.M{"place_id": placeID} }
	cursor, _ := Reviews.Find(context.Background(), filter); var list []bson.M; cursor.All(context.Background(), &list)
	if list == nil { list = []bson.M{} }; jsonResponse(w, 200, list)
}

func GetMyReviews(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := requireLogin(w, r); if !ok { return }
	pipeline := bson.A{bson.M{"$match": bson.M{"user_id": uid}}, bson.M{"$addFields": bson.M{"place_obj_id": bson.M{"$toObjectId": "$place_id"}}}, bson.M{"$lookup": bson.M{"from": "Data_Travel", "localField": "place_obj_id", "foreignField": "_id", "as": "place_info"}}, bson.M{"$unwind": bson.M{"path": "$place_info", "preserveNullAndEmptyArrays": true}}, bson.M{"$project": bson.M{"_id": 1, "rating": 1, "comment": 1, "place_name": "$place_info.name", "place_id": 1}}}
	cursor, _ := Reviews.Aggregate(context.Background(), pipeline); var list []bson.M; cursor.All(context.Background(), &list); if list == nil { list = []bson.M{} }
	jsonResponse(w, 200, list)
}

func DeleteReview(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := requireLogin(w, r); if !ok { return }
	id := strings.TrimPrefix(r.URL.Path, "/api/reviews/"); id = strings.TrimPrefix(id, "/api/my-reviews/")
	objID, _ := primitive.ObjectIDFromHex(id)
	var rev bson.M; if err := Reviews.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&rev); err != nil { jsonResponse(w, 404, map[string]string{"error": "review not found"}); return }
	if role != "admin" && rev["user_id"] != uid { jsonResponse(w, 403, map[string]string{"error": "permission denied"}); return }
	Reviews.DeleteOne(context.Background(), bson.M{"_id": objID}); jsonResponse(w, 200, map[string]string{"message": "review deleted"})
}

func GetPhoto(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/photo/"); objID, err := primitive.ObjectIDFromHex(id)
	if err != nil { jsonResponse(w, 400, map[string]string{"error": "invalid photo id"}); return }
	w.Header().Set("Content-Type", "image/jpeg"); w.Header().Set("Cache-Control", "public, max-age=86400")
	stream, err := photoBucket.OpenDownloadStream(objID)
	if err != nil { jsonResponse(w, 404, map[string]string{"error": "photo not found"}); return }
	defer stream.Close(); io.Copy(w, stream)
}