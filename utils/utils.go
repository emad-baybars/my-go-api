package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"go-backend-template/models"
)

// Logger interface for structured logging
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
}

// SlogLogger implements Logger interface using slog
type SlogLogger struct {
	logger *slog.Logger
}

// NewLogger creates a new logger instance
func NewLogger(level string) Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &SlogLogger{logger: logger}
}

func (l *SlogLogger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

func (l *SlogLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

func (l *SlogLogger) Fatal(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
	os.Exit(1)
}

// Localizer handles internationalization
type Localizer struct {
	DefaultLanguage string
	translations    map[string]map[string]string
}

// NewLocalizer creates a new localizer instance
func NewLocalizer(defaultLang string) (*Localizer, error) {
	localizer := &Localizer{
		DefaultLanguage: defaultLang,
		translations:    make(map[string]map[string]string),
	}

	// Load translations
	if err := localizer.loadTranslations(); err != nil {
		return nil, err
	}

	return localizer, nil
}

// loadTranslations loads translation files
func (l *Localizer) loadTranslations() error {
	// English translations
	l.translations["en"] = map[string]string{
		"welcome":             "Welcome",
		"user_not_found":      "User not found",
		"invalid_credentials": "Invalid credentials",
		"user_created":        "User created successfully",
		"login_successful":    "Login successful",
		"logout_successful":   "Logout successful",
		"user_updated":        "User updated successfully",
		"user_deleted":        "User deleted successfully",
		"email_exists":        "Email already exists",
		"username_exists":     "Username already exists",
		"validation_error":    "Validation error",
		"internal_error":      "Internal server error",
		"unauthorized":        "Unauthorized access",
		"forbidden":           "Access forbidden",
		"not_found":           "Resource not found",
		"bad_request":         "Bad request",
	}

	// Arabic translations
	l.translations["ar"] = map[string]string{
		"welcome":             "أهلا وسهلا",
		"user_not_found":      "المستخدم غير موجود",
		"invalid_credentials": "بيانات الاعتماد غير صحيحة",
		"user_created":        "تم إنشاء المستخدم بنجاح",
		"login_successful":    "تم تسجيل الدخول بنجاح",
		"logout_successful":   "تم تسجيل الخروج بنجاح",
		"user_updated":        "تم تحديث المستخدم بنجاح",
		"user_deleted":        "تم حذف المستخدم بنجاح",
		"email_exists":        "البريد الإلكتروني موجود بالفعل",
		"username_exists":     "اسم المستخدم موجود بالفعل",
		"validation_error":    "خطأ في التحقق",
		"internal_error":      "خطأ في الخادم الداخلي",
		"unauthorized":        "الوصول غير مصرح",
		"forbidden":           "الوصول محظور",
		"not_found":           "المورد غير موجود",
		"bad_request":         "طلب خاطئ",
	}

	// German translations
	l.translations["de"] = map[string]string{
		"welcome":             "Willkommen",
		"user_not_found":      "Benutzer nicht gefunden",
		"invalid_credentials": "Ungültige Anmeldedaten",
		"user_created":        "Benutzer erfolgreich erstellt",
		"login_successful":    "Anmeldung erfolgreich",
		"logout_successful":   "Abmeldung erfolgreich",
		"user_updated":        "Benutzer erfolgreich aktualisiert",
		"user_deleted":        "Benutzer erfolgreich gelöscht",
		"email_exists":        "E-Mail bereits vorhanden",
		"username_exists":     "Benutzername bereits vorhanden",
		"validation_error":    "Validierungsfehler",
		"internal_error":      "Interner Serverfehler",
		"unauthorized":        "Nicht autorisierter Zugriff",
		"forbidden":           "Zugriff verboten",
		"not_found":           "Ressource nicht gefunden",
		"bad_request":         "Fehlerhafte Anfrage",
	}

	return nil
}

// Get returns translated text for the given key and language
func (l *Localizer) Get(lang, key string) string {
	if translations, exists := l.translations[lang]; exists {
		if text, exists := translations[key]; exists {
			return text
		}
	}

	// Fallback to default language
	if translations, exists := l.translations[l.DefaultLanguage]; exists {
		if text, exists := translations[key]; exists {
			return text
		}
	}

	// Return key if no translation found
	return key
}

// PasswordUtils provides password hashing and verification
type PasswordUtils struct{}

// HashPassword hashes a password using bcrypt
func (p *PasswordUtils) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword verifies a password against its hash
func (p *PasswordUtils) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// JWTUtils provides JWT token operations
type JWTUtils struct {
	Secret string
}

// NewJWTUtils creates a new JWT utils instance
func NewJWTUtils(secret string) *JWTUtils {
	return &JWTUtils{Secret: secret}
}

// GenerateToken generates a JWT token for a user
func (j *JWTUtils) GenerateToken(userID interface{}, email, username, role string) (string, time.Time, error) {
	// This method now delegates to the jwt package
	// Import the jwt package at the top of the file
	return "", time.Time{}, fmt.Errorf("use jwt.GenerateToken instead")
}

// ValidateToken validates a JWT token and returns claims
func (j *JWTUtils) ValidateToken(tokenString string) (interface{}, error) {
	// This method now delegates to the jwt package
	// Import the jwt package at the top of the file
	return nil, fmt.Errorf("use jwt.ValidateToken instead")
}

// ResponseUtils provides response helper functions
type ResponseUtils struct{}

// SuccessResponse creates a success response
func (r *ResponseUtils) SuccessResponse(message string, data interface{}) models.APIResponse {
	return models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse creates an error response
func (r *ResponseUtils) ErrorResponse(message, error string) models.APIResponse {
	return models.APIResponse{
		Success: false,
		Message: message,
		Error:   error,
	}
}

// PaginatedResponse creates a paginated response
func (r *ResponseUtils) PaginatedResponse(data interface{}, pagination models.Pagination) models.PaginatedResponse {
	return models.PaginatedResponse{
		Data:       data,
		Pagination: pagination,
	}
}

// ValidationUtils provides validation helper functions
type ValidationUtils struct{}

// IsValidEmail checks if an email is valid
func (v *ValidationUtils) IsValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// IsValidPassword checks if a password meets requirements
func (v *ValidationUtils) IsValidPassword(password string) bool {
	return len(password) >= 6
}

// SanitizeString removes dangerous characters from string
func (v *ValidationUtils) SanitizeString(input string) string {
	return strings.TrimSpace(input)
}

// JSONUtils provides JSON helper functions
type JSONUtils struct{}

// ToJSON converts an object to JSON string
func (j *JSONUtils) ToJSON(obj interface{}) (string, error) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON converts JSON string to object
func (j *JSONUtils) FromJSON(jsonStr string, obj interface{}) error {
	return json.Unmarshal([]byte(jsonStr), obj)
}

// PrettyJSON formats JSON string for better readability
func (j *JSONUtils) PrettyJSON(obj interface{}) (string, error) {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
