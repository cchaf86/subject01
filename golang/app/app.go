package app

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const fixedUserID = "00000000000000000000000000000001"

type User struct {
	ID            string         `json:"id" query:"id" gorm:"type:varchar(32);primary_key;"`
	CreatedAt     time.Time      `json:"created_at" query:"created_at" gorm:"<-:create"`
	UpdatedAt     time.Time      `json:"updated_at" query:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" query:"deleted_at" gorm:"index"`
	CreatedBy     string         `json:"created_by" query:"created_by" gorm:"type:varchar(32);not null"`
	UpdatedBy     string         `json:"updated_by" query:"updated_by" gorm:"type:varchar(32)"`
	DeletedBy     string         `json:"deleted_by" query:"deleted_by" gorm:"type:varchar(32)"`
	FirstName     string         `json:"firstName" query:"first_name" gorm:"type:varchar(64);not null"`
	LastName      string         `json:"lastName" query:"last_name" gorm:"type:varchar(64);not null"`
	Email         string         `json:"email" query:"email" gorm:"type:varchar(255);not null"`
	Phone         string         `json:"phone" query:"phone" gorm:"type:varchar(64);not null"`
	ProfileBase64 string         `json:"profileBase64" query:"profile_base64" gorm:"type:text;not null"`
	BirthDay      string         `json:"birthDay" query:"birth_day" gorm:"type:varchar(64);not null"`
	Occupation    string         `json:"occupation" query:"occupation" gorm:"type:varchar(64);not null"`
	Sex           string         `json:"sex" query:"sex" gorm:"type:varchar(64);not null"`
}

type Server struct {
	DB *gorm.DB
}

func NewServer(db *gorm.DB) *Server {
	return &Server{DB: db}
}

type CreateProfileRequest struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	ProfileBase64 string `json:"profileBase64"`
	BirthDay      string `json:"birthDay"`
	Occupation    string `json:"occupation"`
	Sex           string `json:"sex"`
}

type CreateProfileResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type OccupationResponse struct {
	Items []string `json:"items"`
}

func MustSetupDB(path string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatalf("migrate db: %v", err)
	}
	return db
}

func RegisterRoutes(mux *http.ServeMux, s *Server) {
	mux.HandleFunc("GET /api/occupations", s.handleOccupations)
	mux.HandleFunc("POST /api/profiles", s.handleCreateProfile)
	mux.HandleFunc("OPTIONS /api/occupations", s.handleOptions)
	mux.HandleFunc("OPTIONS /api/profiles", s.handleOptions)
}

func (s *Server) handleOptions(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleOccupations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	resp := OccupationResponse{
		Items: []string{
			"Developer",
			"Tester",
			"System Analyst",
			"Project Manager",
			"Support",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := validateRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := User{
		ID:            newUUID32(),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		Phone:         req.Phone,
		ProfileBase64: req.ProfileBase64,
		BirthDay:      req.BirthDay,
		Occupation:    req.Occupation,
		Sex:           req.Sex,
		CreatedBy:     fixedUserID,
	}

	if err := s.DB.Create(&user).Error; err != nil {
		log.Printf("insert user: %v", err)
		http.Error(w, "failed to save", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateProfileResponse{
		ID:      user.ID,
		Message: "save data success",
	})
}

func validateRequest(req CreateProfileRequest) error {
	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Phone == "" || req.ProfileBase64 == "" || req.BirthDay == "" || req.Occupation == "" || req.Sex == "" {
		return errors.New("all fields are required")
	}

	digitsOnly := regexp.MustCompile(`^[0-9]+$`)
	if !digitsOnly.MatchString(req.Phone) {
		return errors.New("phone must contain digits only")
	}

	if _, err := time.Parse("02/01/2006", req.BirthDay); err != nil {
		return errors.New("birthDay must be in format DD/MM/YYYY")
	}

	return nil
}

// newUUID32 generates a UUID v4 without dashes (32 chars).
func newUUID32() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
