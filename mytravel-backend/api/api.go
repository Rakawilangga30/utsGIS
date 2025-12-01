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
	"sync" // [PENTING] Untuk memastikan DB hanya konek sekali
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
   GLOBAL VARS
======================================================= */

var (
	mongoDB     *mongo.Database
	photoBucket *gridfs.Bucket
	Store       *sessions.CookieStore
	Users       *mongo.Collection
	Places      *mongo.Collection
	Reviews     *mongo.Collection
	dbOnce      sync.Once // Singleton pattern untuk DB
)

/* =======================================================
   KONEKSI DATABASE (Lazy Connection)
======================================================= */

// ConnectDB memastikan koneksi hanya dibuat sekali (aman untuk Vercel & Local)
func ConnectDB() {
	dbOnce.Do(func() {
		// Load .env tapi abaikan error (karena di Vercel tidak ada file .env fisik)
		_ = godotenv.Load()

		uri := os.Getenv("MONGO_URI")
		dbName := os.Getenv("DB_NAME")
		secret := os.Getenv("SESSION_KEY")

		// Fallback default value
		if dbName == "" {
			dbName = "MyTravelDB"
		}
		if secret == "" {
			secret = "MYTRAVEL_SECRET_KEY_GANTI_INI"
		}

		if uri == "" {
			fmt.Println("Warning: MONGO_URI kosong!")
		}

		// Setup Session
		Store = sessions.NewCookieStore([]byte(secret))
		Store.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   3600 * 24, // 1 hari
			HttpOnly: true,
			// Penting: SameSiteNoneMode diperlukan jika frontend & backend beda domain/port
			SameSite: http.SameSiteNoneMode,
			Secure:   true, // Harus true jika menggunakan SameSiteNoneMode
		}

		// Setup Mongo
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
	})
}

/* =======================================================
   HELPERS & CORS (FIXED)
======================================================= */

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// EnableCORS diperbaiki untuk mendukung Cookie/Login
func EnableCORS(w http.ResponseWriter, r *http.Request) {
	// Ambil Origin dari request (misal: http://127.0.0.1:5500)
	origin := r.Header.Get("Origin")
	
	// Jika ada origin, izinkan spesifik (bukan *) agar Credentials bisa true
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true") // Wajib true agar session login tersimpan
}

func getSession(r *http.Request) (*sessions.Session, error) {
	// Pastikan Store sudah diinit sebelum dipanggil
	if Store == nil {
		ConnectDB()
	}
	return Store.Get(r, "mytravel-session")
}

func requireLogin(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	sess, _ := getSession(r)
	id, ok1 := sess.Values["user_id"].(string)
	role, ok2 := sess.Values["role"].(string)

	if !ok1 || !ok2 {
		jsonResponse(w, 401, map[string]string{"error": "unauthorized - silakan login"})
		return "", "", false
	}
	return id, role, true
}

/* =======================================================
   MAIN ROUTER (Entry Point)
======================================================= */

func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pastikan DB connect setiap ada request
		ConnectDB()

		EnableCORS(w, r)
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
		// Handle ID di URL (e.g., /api/places/123)
		if strings.HasPrefix(p, "/api/places/") {
			if r.Method == "PUT" {
				UpdatePlace(w, r); return
			}
			if r.Method == "DELETE" {
				DeletePlace(w, r); return
			}
		}

		/* -------- MY PLACES (Dashboard) -------- */
		if p == "/api/my-places" && r.Method == "GET" {
			GetMyPlaces(w, r); return
		}
		if strings.HasPrefix(p, "/api/my-places/") && r.Method == "DELETE" {
			DeletePlace(w, r); return
		}

		/* -------- REVIEWS -------- */
		if p == "/api/reviews" && r.Method == "GET" {
			GetReviews(w, r); return
		}
		if p == "/api/reviews" && r.Method == "POST" {
			CreateReview(w, r); return
		}
		
		// Dashboard My Reviews
		if p == "/api/my-reviews" && r.Method == "GET" {
			GetMyReviews(w, r); return
		}
		if strings.HasPrefix(p, "/api/my-reviews/") && r.Method == "DELETE" {
			DeleteReview(w, r); return
		}
		
		if strings.HasPrefix(p, "/api/reviews/") && r.Method == "DELETE" {
			DeleteReview(w, r); return
		}

		/* -------- PHOTO -------- */
		if strings.HasPrefix(p, "/api/photo/") {
			GetPhoto(w, r); return
		}

		jsonResponse(w, 404, map[string]string{"error": "route not found: " + p})
	}
}

/* =======================================================
   FUNGSI HANDLER (CRUD)
======================================================= */

// --- AUTH ---

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
	if !ok { return }

	jsonResponse(w, 200, map[string]string{
		"user_id": uid,
		"role":    role,
	})
}

// --- PLACES ---

func CreatePlace(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := requireLogin(w, r)
	if !ok { return }

	r.ParseMultipartForm(10 << 20)

	name := r.FormValue("name")
	category := r.FormValue("category")
	description := r.FormValue("description")
	address := r.FormValue("address")
	latStr := r.FormValue("lat")
	lngStr := r.FormValue("lng")

	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)

	file, header, err := r.FormFile("photo")
	photoID := ""
	
	if err == nil {
		defer file.Close()
		uploadStream, _ := photoBucket.OpenUploadStream(header.Filename)
		io.Copy(uploadStream, file)
		uploadStream.Close()
		photoID = uploadStream.FileID.(primitive.ObjectID).Hex()
	}

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
	if places == nil { places = []bson.M{} }
	jsonResponse(w, 200, places)
}

func GetMyPlaces(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := requireLogin(w, r)
	if !ok { return }

	cursor, _ := Places.Find(context.Background(), bson.M{"created_by": uid})
	var places []bson.M
	cursor.All(context.Background(), &places)
	if places == nil { places = []bson.M{} }
	jsonResponse(w, 200, places)
}

func UpdatePlace(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := requireLogin(w, r)
	if !ok { return }

	id := strings.TrimPrefix(r.URL.Path, "/api/places/")
	objID, _ := primitive.ObjectIDFromHex(id)

	var old bson.M
	err := Places.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&old)
	if err != nil {
		jsonResponse(w, 404, map[string]string{"error": "place not found"})
		return
	}

	if role != "admin" && old["created_by"] != uid {
		jsonResponse(w, 403, map[string]string{"error": "permission denied"})
		return
	}

	r.ParseMultipartForm(10 << 20)
	
	// Update data dasar
	updateData := bson.M{
		"name":        r.FormValue("name"),
		"category":    r.FormValue("category"),
		"description": r.FormValue("description"),
		"address":     r.FormValue("address"),
		"lat":         0.0,
		"lng":         0.0,
	}
	
	if val := r.FormValue("lat"); val != "" {
	    f, _ := strconv.ParseFloat(val, 64)
	    updateData["lat"] = f
	} else { updateData["lat"] = old["lat"] }
	
	if val := r.FormValue("lng"); val != "" {
	    f, _ := strconv.ParseFloat(val, 64)
	    updateData["lng"] = f
	} else { updateData["lng"] = old["lng"] }

	// Cek foto baru
	file, header, err := r.FormFile("photo")
	if err == nil {
		defer file.Close()
		upload, _ := photoBucket.OpenUploadStream(header.Filename)
		io.Copy(upload, file)
		upload.Close()
		updateData["photo_id"] = upload.FileID.(primitive.ObjectID).Hex()
	}

	Places.UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$set": updateData})
	jsonResponse(w, 200, map[string]string{"message": "place updated"})
}

func DeletePlace(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := requireLogin(w, r)
	if !ok { return }

	// Handle dua jenis URL (api/places/ID atau api/my-places/ID)
	id := strings.TrimPrefix(r.URL.Path, "/api/places/")
	id = strings.TrimPrefix(id, "/api/my-places/")
	
	objID, _ := primitive.ObjectIDFromHex(id)

	var place bson.M
	if err := Places.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&place); err != nil {
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

// --- REVIEWS ---

func CreateReview(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := requireLogin(w, r)
	if !ok { return }

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
	
	// Jika tidak ada parameter place_id, ambil semua (untuk rekomendasi)
	filter := bson.M{}
	if placeID != "" {
	    filter = bson.M{"place_id": placeID}
	}

	cursor, _ := Reviews.Find(context.Background(), filter)
	var list []bson.M
	cursor.All(context.Background(), &list)
	if list == nil { list = []bson.M{} }
	
	jsonResponse(w, 200, list)
}

func GetMyReviews(w http.ResponseWriter, r *http.Request) {
    uid, _, ok := requireLogin(w, r)
    if !ok { return }
    
    // Agregation pipeline untuk join dengan nama tempat
    pipeline := bson.A{
        bson.M{"$match": bson.M{"user_id": uid}},
        // Konversi place_id (string) ke ObjectId agar bisa di-lookup
        bson.M{"$addFields": bson.M{
            "place_obj_id": bson.M{"$toObjectId": "$place_id"},
        }},
        bson.M{"$lookup": bson.M{
            "from": "Data_Travel",
            "localField": "place_obj_id",
            "foreignField": "_id",
            "as": "place_info",
        }},
        bson.M{"$unwind": bson.M{
            "path": "$place_info", 
            "preserveNullAndEmptyArrays": true, // Biar review tetap muncul walau tempat dihapus
        }},
        bson.M{"$project": bson.M{
            "_id": 1, "rating": 1, "comment": 1,
            "place_name": "$place_info.name",
            "place_id": 1,
        }},
    }
    
    cursor, err := Reviews.Aggregate(context.Background(), pipeline)
    if err != nil {
        jsonResponse(w, 500, map[string]string{"error": err.Error()})
        return
    }
    
    var list []bson.M
    cursor.All(context.Background(), &list)
    if list == nil { list = []bson.M{} }
    
    jsonResponse(w, 200, list)
}


func DeleteReview(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := requireLogin(w, r)
	if !ok { return }

	id := strings.TrimPrefix(r.URL.Path, "/api/reviews/")
	id = strings.TrimPrefix(id, "/api/my-reviews/")
	objID, _ := primitive.ObjectIDFromHex(id)

	var rev bson.M
	if err := Reviews.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&rev); err != nil {
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

// --- PHOTO ---

func GetPhoto(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/photo/")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid photo id"})
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	// Cache control biar gambar tidak diload ulang terus
	w.Header().Set("Cache-Control", "public, max-age=86400") 

	stream, err := photoBucket.OpenDownloadStream(objID)
	if err != nil {
		jsonResponse(w, 404, map[string]string{"error": "photo not found"})
		return
	}
	defer stream.Close()
	io.Copy(w, stream)
}