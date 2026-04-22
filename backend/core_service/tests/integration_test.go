//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Artem09076/dp/backend/core_service/internal/application/discounts"
	"github.com/Artem09076/dp/backend/core_service/internal/application/profile"
	"github.com/Artem09076/dp/backend/core_service/internal/application/reviews"
	"github.com/Artem09076/dp/backend/core_service/internal/application/services"
	"github.com/Artem09076/dp/backend/core_service/internal/config"
	"github.com/Artem09076/dp/backend/core_service/internal/lib/jwt"
	discountdto "github.com/Artem09076/dp/backend/core_service/internal/presentation/discounts/dto"
	discounthandlers "github.com/Artem09076/dp/backend/core_service/internal/presentation/discounts/handlers"
	coremiddleware "github.com/Artem09076/dp/backend/core_service/internal/presentation/middleware"
	profiledto "github.com/Artem09076/dp/backend/core_service/internal/presentation/profile/dto"
	profilehandlers "github.com/Artem09076/dp/backend/core_service/internal/presentation/profile/handlers"
	reviewdto "github.com/Artem09076/dp/backend/core_service/internal/presentation/reviews/dto"
	reviewhandlers "github.com/Artem09076/dp/backend/core_service/internal/presentation/reviews/handlers"
	servicedto "github.com/Artem09076/dp/backend/core_service/internal/presentation/services/dto"
	servicehandlers "github.com/Artem09076/dp/backend/core_service/internal/presentation/services/handlers"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/core_service/internal/storage/rabbit"
	redisstorage "github.com/Artem09076/dp/backend/core_service/internal/storage/redis"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB           *sql.DB
	testRedisClient  *redis.Client
	testRedisWrapper *redisstorage.RedisClient
	testRabbitConn   *amqp.Connection
	testRabbitCh     *amqp.Channel
	testRouter       *chi.Mux
	testQueries      *sqlc.Queries
	testLogger       *slog.Logger
	testConfig       *config.Config

	// Test UUIDs
	TestServiceID    = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	TestDiscountID   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	TestPerformerID  = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	TestClientID     = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	TestAdminID      = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	TestUnverifiedID = uuid.MustParse("66666666-6666-6666-6666-666666666666")
	TestReviewID     = uuid.MustParse("77777777-7777-7777-7777-777777777777")
	TestBookingID    = uuid.MustParse("88888888-8888-8888-8888-888888888888")

	jwtSecret = "test-secret-key-for-core-service-integration-tests"
)

func TestMain(m *testing.M) {
	testLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	pool, err := dockertest.NewPool("")
	if err != nil {
		testLogger.Error("Could not connect to docker", "error", err)
		os.Exit(1)
	}

	// Start PostgreSQL
	postgresResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15",
		Env: []string{
			"POSTGRES_PASSWORD=test",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=testdb",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		testLogger.Error("Could not start postgres", "error", err)
		os.Exit(1)
	}

	// Start Redis
	redisResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "7-alpine",
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
	})
	if err != nil {
		testLogger.Error("Could not start redis", "error", err)
		os.Exit(1)
	}

	// Start RabbitMQ
	rabbitResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "rabbitmq",
		Tag:        "3-management-alpine",
		Env: []string{
			"RABBITMQ_DEFAULT_USER=guest",
			"RABBITMQ_DEFAULT_PASS=guest",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
	})
	if err != nil {
		testLogger.Error("Could not start rabbitmq", "error", err)
		os.Exit(1)
	}

	// Connect to PostgreSQL
	postgresPort := postgresResource.GetPort("5432/tcp")
	testDB, err = sql.Open("postgres", fmt.Sprintf("postgres://postgres:test@localhost:%s/testdb?sslmode=disable", postgresPort))
	if err != nil {
		testLogger.Error("Could not connect to postgres", "error", err)
		os.Exit(1)
	}

	err = pool.Retry(func() error {
		return testDB.Ping()
	})
	if err != nil {
		testLogger.Error("Could not ping postgres", "error", err)
		os.Exit(1)
	}

	// Run migrations
	if err := runMigrations(testDB); err != nil {
		testLogger.Error("Could not run migrations", "error", err)
		os.Exit(1)
	}

	// Seed test data
	if err := seedTestData(testDB); err != nil {
		testLogger.Error("Could not seed test data", "error", err)
		os.Exit(1)
	}

	// Connect to Redis
	redisPort := redisResource.GetPort("6379/tcp")
	testRedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:" + redisPort,
		DB:   1,
	})

	err = pool.Retry(func() error {
		return testRedisClient.Ping(context.Background()).Err()
	})
	if err != nil {
		testLogger.Error("Could not connect to redis", "error", err)
		os.Exit(1)
	}

	// Wrap Redis client
	testRedisWrapper, err = redisstorage.NewRedisClient(testLogger, "localhost:"+redisPort, "", 1)
	if err != nil {
		testLogger.Error("Could not create redis wrapper", "error", err)
		os.Exit(1)
	}

	// Connect to RabbitMQ
	rabbitPort := rabbitResource.GetPort("5672/tcp")
	testRabbitConn, err = amqp.Dial(fmt.Sprintf("amqp://guest:guest@localhost:%s/", rabbitPort))
	if err != nil {
		testLogger.Error("Could not connect to rabbitmq", "error", err)
		os.Exit(1)
	}

	testRabbitCh, err = testRabbitConn.Channel()
	if err != nil {
		testLogger.Error("Could not create rabbitmq channel", "error", err)
		os.Exit(1)
	}

	// Declare queue
	_, err = testRabbitCh.QueueDeclare("profile_queue", true, false, false, false, nil)
	if err != nil {
		testLogger.Error("Could not declare queue", "error", err)
		os.Exit(1)
	}

	// Setup router
	testRouter = setupTestRouter(testDB, testRedisWrapper, testRabbitCh)
	testQueries = sqlc.New(testDB)

	// Run tests
	code := m.Run()

	// Cleanup
	if testRabbitCh != nil {
		testRabbitCh.Close()
	}
	if testRabbitConn != nil {
		testRabbitConn.Close()
	}
	if testRedisWrapper != nil {
		testRedisWrapper.Close()
	}
	if testRedisClient != nil {
		testRedisClient.Close()
	}
	if testDB != nil {
		testDB.Close()
	}

	// Purge resources
	if err := pool.Purge(postgresResource); err != nil {
		testLogger.Error("Failed to purge postgres resource", "error", err)
	}
	if err := pool.Purge(redisResource); err != nil {
		testLogger.Error("Failed to purge redis resource", "error", err)
	}
	if err := pool.Purge(rabbitResource); err != nil {
		testLogger.Error("Failed to purge rabbitmq resource", "error", err)
	}

	os.Exit(code)
}

func runMigrations(db *sql.DB) error {
	// Drop types if exist
	dropTypes := []string{
		`DROP TYPE IF EXISTS user_role CASCADE`,
		`DROP TYPE IF EXISTS business_type CASCADE`,
		`DROP TYPE IF EXISTS verification_status CASCADE`,
		`DROP TYPE IF EXISTS booking_status CASCADE`,
		`DROP TYPE IF EXISTS discoutn_type CASCADE`,
	}

	for _, migration := range dropTypes {
		if _, err := db.Exec(migration); err != nil {
			testLogger.Warn("Failed to drop type (may not exist)", "error", err)
		}
	}

	// Create enum types
	enumTypes := []string{
		`CREATE TYPE user_role AS ENUM ('admin', 'client', 'performer')`,
		`CREATE TYPE business_type AS ENUM ('IP', 'self_employed')`,
		`CREATE TYPE verification_status AS ENUM ('pending', 'verified', 'rejected')`,
		`CREATE TYPE booking_status AS ENUM ('pending', 'confirmed', 'completed', 'cancelled')`,
		`CREATE TYPE discoutn_type AS ENUM ('percentage', 'fixed_amount')`,
	}

	for _, migration := range enumTypes {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("enum migration failed: %s, error: %w", migration, err)
		}
	}

	// Drop tables if exist
	dropTables := []string{
		`DROP TABLE IF EXISTS reviews CASCADE`,
		`DROP TABLE IF EXISTS bookings CASCADE`,
		`DROP TABLE IF EXISTS discounts CASCADE`,
		`DROP TABLE IF EXISTS services CASCADE`,
		`DROP TABLE IF EXISTS users CASCADE`,
	}

	for _, migration := range dropTables {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("drop table failed: %s, error: %w", migration, err)
		}
	}

	// Create tables
	tables := []string{
		`CREATE TABLE users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash BYTEA NOT NULL,
			role user_role NOT NULL,
			inn VARCHAR(12) UNIQUE,
			business_type business_type,
			verification_status verification_status NOT NULL DEFAULT 'pending',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE services (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			performer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			price BIGINT NOT NULL,
			duration_minutes INT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE discounts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
			type discoutn_type NOT NULL,
			value INT NOT NULL,
			valid_from TIMESTAMPTZ NOT NULL,
			valid_to TIMESTAMPTZ NOT NULL,
			max_uses INT NOT NULL,
			used_count INT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE bookings (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
			base_price INT NOT NULL,
			discount_id UUID REFERENCES discounts(id) ON DELETE SET NULL,
			final_price INT NOT NULL,
			booking_time TIMESTAMPTZ NOT NULL,
			status booking_status NOT NULL DEFAULT 'pending',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE reviews (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
			rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
			comment TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
	}

	for i, migration := range tables {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("table migration %d failed: %s, error: %w", i, migration, err)
		}
	}

	// Create indexes
	indexes := []string{
		`CREATE INDEX idx_users_email ON users(email)`,
		`CREATE INDEX idx_services_performer ON services(performer_id)`,
		`CREATE INDEX idx_bookings_client ON bookings(client_id)`,
		`CREATE INDEX idx_bookings_service ON bookings(service_id)`,
	}

	for _, migration := range indexes {
		if _, err := db.Exec(migration); err != nil {
			testLogger.Warn("Failed to create index (may already exist)", "error", err)
		}
	}

	// Create update timestamp function
	_, err := db.Exec(`
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
		   NEW.updated_at = NOW();
		   RETURN NEW;
		END;
		$$ language 'plpgsql'
	`)
	if err != nil {
		return fmt.Errorf("failed to create function: %w", err)
	}

	// Create triggers
	triggers := []string{
		`CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
		`CREATE TRIGGER update_services_updated_at BEFORE UPDATE ON services
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
		`CREATE TRIGGER update_bookings_updated_at BEFORE UPDATE ON bookings
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
	}

	for _, migration := range triggers {
		if _, err := db.Exec(migration); err != nil {
			testLogger.Warn("Failed to create trigger (may already exist)", "error", err)
		}
	}

	return nil
}

func seedTestData(db *sql.DB) error {
	passwordHash := []byte("$2a$10$test_hash_for_integration_tests")

	// Insert users
	users := []struct {
		id       uuid.UUID
		email    string
		name     string
		role     string
		verified string
	}{
		{TestClientID, "client@test.com", "Test Client", "client", "verified"},
		{TestPerformerID, "performer@test.com", "Test Performer", "performer", "verified"},
		{TestAdminID, "admin@test.com", "Test Admin", "admin", "verified"},
		{TestUnverifiedID, "unverified@test.com", "Unverified Performer", "performer", "pending"},
	}

	for _, u := range users {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", u.id).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check user: %w", err)
		}
		if count == 0 {
			_, err = db.Exec(`
				INSERT INTO users (id, email, name, password_hash, role, verification_status)
				VALUES ($1, $2, $3, $4, $5, $6)`,
				u.id, u.email, u.name, passwordHash, u.role, u.verified)
			if err != nil {
				return fmt.Errorf("failed to insert user %s: %w", u.email, err)
			}
		}
	}

	// Insert service
	var serviceCount int
	err := db.QueryRow("SELECT COUNT(*) FROM services WHERE id = $1", TestServiceID).Scan(&serviceCount)
	if err != nil {
		return fmt.Errorf("failed to check service: %w", err)
	}
	if serviceCount == 0 {
		_, err = db.Exec(`
			INSERT INTO services (id, title, description, price, duration_minutes, performer_id)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			TestServiceID, "Test Service", "Test Description", 1000, 60, TestPerformerID)
		if err != nil {
			return fmt.Errorf("failed to insert service: %w", err)
		}
	}

	// Insert discount
	var discountCount int
	err = db.QueryRow("SELECT COUNT(*) FROM discounts WHERE id = $1", TestDiscountID).Scan(&discountCount)
	if err != nil {
		return fmt.Errorf("failed to check discount: %w", err)
	}
	if discountCount == 0 {
		_, err = db.Exec(`
			INSERT INTO discounts (id, service_id, type, value, max_uses, valid_from, valid_to)
			VALUES ($1, $2, 'percentage', 20, 100, NOW() - INTERVAL '1 day', NOW() + INTERVAL '30 days')`,
			TestDiscountID, TestServiceID)
		if err != nil {
			return fmt.Errorf("failed to insert discount: %w", err)
		}
	}

	// Insert booking
	var bookingCount int
	err = db.QueryRow("SELECT COUNT(*) FROM bookings WHERE id = $1", TestBookingID).Scan(&bookingCount)
	if err != nil {
		return fmt.Errorf("failed to check booking: %w", err)
	}
	if bookingCount == 0 {
		_, err = db.Exec(`
			INSERT INTO bookings (id, client_id, service_id, base_price, final_price, booking_time, status)
			VALUES ($1, $2, $3, 1000, 1000, $4, 'completed')`,
			TestBookingID, TestClientID, TestServiceID, time.Now().Add(-24*time.Hour))
		if err != nil {
			return fmt.Errorf("failed to insert booking: %w", err)
		}
	}

	// Insert review
	var reviewCount int
	err = db.QueryRow("SELECT COUNT(*) FROM reviews WHERE id = $1", TestReviewID).Scan(&reviewCount)
	if err != nil {
		return fmt.Errorf("failed to check review: %w", err)
	}
	if reviewCount == 0 {
		_, err = db.Exec(`
			INSERT INTO reviews (id, booking_id, rating, comment)
			VALUES ($1, $2, 5, 'Great service!')`,
			TestReviewID, TestBookingID)
		if err != nil {
			return fmt.Errorf("failed to insert review: %w", err)
		}
	}

	testLogger.Info("Test data seeded successfully",
		"service_id", TestServiceID,
		"discount_id", TestDiscountID,
		"booking_id", TestBookingID,
		"review_id", TestReviewID)

	return nil
}

func setupTestRouter(db *sql.DB, redisClient *redisstorage.RedisClient, rabbitCh *amqp.Channel) *chi.Mux {
	log := testLogger
	queries := sqlc.New(db)
	validator := jwt.NewValidator(jwtSecret, redisClient)
	publisher := rabbit.NewPublisher(rabbitCh)

	// Initialize services
	profileService := profile.NewProfileService(queries, log, publisher)
	profileHandlers := profilehandlers.NewProfileHandler(profileService, log)

	serviceService := services.NewService(queries, log)
	serviceHandlers := servicehandlers.NewServiceHandler(serviceService, log)

	discountService := discounts.NewDiscountService(queries, log)
	discountHandlers := discounthandlers.NewDiscountsHandler(discountService, log)

	reviewService := reviews.NewReviewService(queries, log)
	reviewHandlers := reviewhandlers.NewReviewHandler(reviewService, log)

	router := chi.NewRouter()
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)

	// Public endpoints (with JWT)
	router.Group(func(r chi.Router) {
		r.Use(coremiddleware.NewJWTMiddleware(log, validator))
		r.Get("/api/v1/profile", profileHandlers.GetProfile())
		r.Patch("/api/v1/profile", profileHandlers.PatchProfile())
		r.Delete("/api/v1/profile", profileHandlers.DeleteProfile())
		r.Get("/api/v1/services", serviceHandlers.SearchServices())
		r.Get("/api/v1/services/{id}", serviceHandlers.GetService())
		r.Get("/api/v1/discounts/{id}", discountHandlers.GetDiscount())
		r.Post("/api/v1/reviews", reviewHandlers.CreateReview())
		r.Get("/api/v1/booking/{bookingID}/reviews", reviewHandlers.GetReviewByBookingID())
		r.Get("/api/v1/reviews/{reviewID}", reviewHandlers.GetReviewByID())
		r.Get("/api/v1/service/{serviceID}/reviews", reviewHandlers.GetReviewsByServiceID())
		r.Patch("/api/v1/reviews/{reviewID}", reviewHandlers.PatchReview())
		r.Delete("/api/v1/reviews/{reviewID}", reviewHandlers.DeleteReview())
	})

	// Endpoints for admin and performer
	router.Group(func(r chi.Router) {
		r.Use(coremiddleware.NewJWTMiddleware(log, validator))
		r.Use(coremiddleware.NewRoleMiddleware(log, []string{"admin", "performer"}))
		r.Post("/api/v1/services", serviceHandlers.CreateService())
		r.Patch("/api/v1/services/{id}", serviceHandlers.PatchService())
		r.Delete("/api/v1/services/{id}", serviceHandlers.DeleteService())
		r.Get("/api/v1/services", serviceHandlers.GetServices())
		r.Post("/api/v1/services/{id}/discounts", discountHandlers.CreateDiscount())
		r.Patch("/api/v1/services/{serviceID}/discounts/{id}", discountHandlers.UpdateDiscount())
		r.Delete("/api/v1/services/{serviceID}/discounts/{id}", discountHandlers.DeleteDiscount())
	})

	// Endpoints only for admin
	router.Group(func(r chi.Router) {
		r.Use(coremiddleware.NewJWTMiddleware(log, validator))
		r.Use(coremiddleware.NewRoleMiddleware(log, []string{"admin"}))
		r.Patch("/api/v1/profile/verification_status", profileHandlers.UpdateVerificationStatus())
	})

	return router
}

func GenerateTestToken(userID uuid.UUID, role string) string {
	claims := jwt2.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt2.NewWithClaims(jwt2.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwtSecret))
	return tokenString
}

func CleanupTestData(t *testing.T) {
	_, err := testDB.Exec("DELETE FROM reviews")
	require.NoError(t, err)
	_, err = testDB.Exec("DELETE FROM bookings")
	require.NoError(t, err)
	_, err = testDB.Exec("DELETE FROM discounts")
	require.NoError(t, err)
	_, err = testDB.Exec("DELETE FROM services")
	require.NoError(t, err)

	if testRedisClient != nil {
		err = testRedisClient.FlushAll(context.Background()).Err()
		require.NoError(t, err)
	}
}

func SkipIfNoDocker(t *testing.T) {
	if os.Getenv("SKIP_DOCKER_TESTS") == "1" {
		t.Skip("Skipping Docker-dependent test")
	}
}

// ========== PROFILE TESTS ==========

func TestProfileGet(t *testing.T) {
	SkipIfNoDocker(t)

	tests := []struct {
		name           string
		userID         uuid.UUID
		role           string
		expectedStatus int
	}{
		{
			name:           "Success - Client",
			userID:         TestClientID,
			role:           "client",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Success - Performer",
			userID:         TestPerformerID,
			role:           "performer",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Success - Admin",
			userID:         TestAdminID,
			role:           "admin",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := GenerateTestToken(tt.userID, tt.role)
			req := httptest.NewRequest("GET", "/api/v1/profile", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			res := httptest.NewRecorder()

			testRouter.ServeHTTP(res, req)

			assert.Equal(t, tt.expectedStatus, res.Code)

			if tt.expectedStatus == http.StatusOK {
				var response profiledto.ProfileResponse
				err := json.NewDecoder(res.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Name)
				assert.NotEmpty(t, response.Email)
				assert.Equal(t, tt.role, response.Role)
			}
		})
	}
}

func TestProfileUpdate(t *testing.T) {
	SkipIfNoDocker(t)
	defer CleanupTestData(t)

	tests := []struct {
		name           string
		userID         uuid.UUID
		role           string
		requestBody    profiledto.PatchProfileRequest
		expectedStatus int
	}{
		{
			name:   "Success - Update name and email",
			userID: TestClientID,
			role:   "client",
			requestBody: profiledto.PatchProfileRequest{
				Name:  "Updated Client Name",
				Email: "updated@test.com",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Success - Update only name",
			userID: TestPerformerID,
			role:   "performer",
			requestBody: profiledto.PatchProfileRequest{
				Name: "New Performer Name",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Success - Update only email",
			userID: TestAdminID,
			role:   "admin",
			requestBody: profiledto.PatchProfileRequest{
				Email: "admin_new@test.com",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := GenerateTestToken(tt.userID, tt.role)
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PATCH", "/api/v1/profile", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			res := httptest.NewRecorder()

			testRouter.ServeHTTP(res, req)

			assert.Equal(t, tt.expectedStatus, res.Code)
		})
	}
}

func TestProfileDelete(t *testing.T) {
	SkipIfNoDocker(t)

	// Create temporary user for deletion test
	tempUserID := uuid.New()
	_, err := testDB.Exec(`
		INSERT INTO users (id, name, email, password_hash, role, verification_status)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		tempUserID, "Temp User", "temp@test.com", []byte("hash"), "client", "verified")
	require.NoError(t, err)

	token := GenerateTestToken(tempUserID, "client")
	req := httptest.NewRequest("DELETE", "/api/v1/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	testRouter.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)

	// Verify user was deleted
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", tempUserID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// ========== SERVICE TESTS ==========

func TestCreateService(t *testing.T) {
	SkipIfNoDocker(t)
	defer CleanupTestData(t)

	tests := []struct {
		name           string
		userID         uuid.UUID
		role           string
		requestBody    servicedto.CreateServiceRequest
		expectedStatus int
	}{
		{
			name:   "Success - Verified performer creates service",
			userID: TestPerformerID,
			role:   "performer",
			requestBody: servicedto.CreateServiceRequest{
				Title:           "New Test Service",
				Description:     "Service Description",
				Price:           1500,
				DurationMinutes: 90,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Forbidden - Unverified performer creates service",
			userID: TestUnverifiedID,
			role:   "performer",
			requestBody: servicedto.CreateServiceRequest{
				Title:           "Unverified Service",
				Description:     "Description",
				Price:           1000,
				DurationMinutes: 60,
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "Forbidden - Client creates service",
			userID: TestClientID,
			role:   "client",
			requestBody: servicedto.CreateServiceRequest{
				Title:           "Client Service",
				Description:     "Description",
				Price:           1000,
				DurationMinutes: 60,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := GenerateTestToken(tt.userID, tt.role)
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/services", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			res := httptest.NewRecorder()

			testRouter.ServeHTTP(res, req)

			assert.Equal(t, tt.expectedStatus, res.Code)

			if tt.expectedStatus == http.StatusOK {
				var response servicedto.ServiceResponse
				err := json.NewDecoder(res.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody.Title, response.Title)
				assert.Equal(t, int64(tt.requestBody.Price), response.Price)
			}
		})
	}
}

func TestGetService(t *testing.T) {
	SkipIfNoDocker(t)

	token := GenerateTestToken(TestClientID, "client")
	req := httptest.NewRequest("GET", "/api/v1/services/"+TestServiceID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	testRouter.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)

	var response servicedto.ServiceResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, TestServiceID.String(), response.ID)
	assert.Equal(t, "Test Service", response.Title)
}

func TestSearchServices(t *testing.T) {
	SkipIfNoDocker(t)

	tests := []struct {
		name           string
		query          string
		page           string
		limit          string
		expectedStatus int
	}{
		{
			name:           "Success - Search by query",
			query:          "Test",
			page:           "1",
			limit:          "10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Success - Empty query",
			query:          "",
			page:           "1",
			limit:          "10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Bad Request - Invalid page",
			query:          "test",
			page:           "invalid",
			limit:          "10",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Bad Request - Invalid limit",
			query:          "test",
			page:           "1",
			limit:          "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := GenerateTestToken(TestClientID, "client")
			url := fmt.Sprintf("/api/v1/services?query=%s&page=%s&limit=%s", tt.query, tt.page, tt.limit)
			req := httptest.NewRequest("GET", url, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			res := httptest.NewRecorder()

			testRouter.ServeHTTP(res, req)

			assert.Equal(t, tt.expectedStatus, res.Code)

			if tt.expectedStatus == http.StatusOK {
				var services []servicedto.ServiceResponse
				err := json.NewDecoder(res.Body).Decode(&services)
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateService(t *testing.T) {
	SkipIfNoDocker(t)
	defer CleanupTestData(t)

	// Create service for update
	serviceID := uuid.New()
	_, err := testDB.Exec(`
		INSERT INTO services (id, performer_id, title, description, price, duration_minutes)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		serviceID, TestPerformerID, "Original Title", "Original Desc", 1000, 60)
	require.NoError(t, err)

	newTitle := "Updated Title"
	newPrice := 2000
	requestBody := servicedto.PatchServiceRequest{
		Title: &newTitle,
		Price: &newPrice,
	}

	token := GenerateTestToken(TestPerformerID, "performer")
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PATCH", "/api/v1/services/"+serviceID.String(), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	testRouter.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)

	// Verify update
	var title string
	var price int64
	err = testDB.QueryRow("SELECT title, price FROM services WHERE id = $1", serviceID).Scan(&title, &price)
	assert.NoError(t, err)
	assert.Equal(t, newTitle, title)
	assert.Equal(t, int64(newPrice), price)
}

func TestDeleteService(t *testing.T) {
	SkipIfNoDocker(t)
	defer CleanupTestData(t)

	// Create service for deletion
	serviceID := uuid.New()
	_, err := testDB.Exec(`
		INSERT INTO services (id, performer_id, title, description, price, duration_minutes)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		serviceID, TestPerformerID, "To Delete", "Desc", 1000, 60)
	require.NoError(t, err)

	token := GenerateTestToken(TestPerformerID, "performer")
	req := httptest.NewRequest("DELETE", "/api/v1/services/"+serviceID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	testRouter.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)

	// Verify deletion
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM services WHERE id = $1", serviceID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// ========== DISCOUNT TESTS ==========

func TestCreateDiscount(t *testing.T) {
	SkipIfNoDocker(t)
	defer CleanupTestData(t)

	// Create service for discount
	serviceID := uuid.New()
	_, err := testDB.Exec(`
		INSERT INTO services (id, performer_id, title, description, price, duration_minutes)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		serviceID, TestPerformerID, "Discount Service", "Desc", 1000, 60)
	require.NoError(t, err)

	tests := []struct {
		name           string
		userID         uuid.UUID
		role           string
		requestBody    discountdto.CreateDiscountRequest
		expectedStatus int
	}{
		{
			name:   "Success - Create percentage discount",
			userID: TestPerformerID,
			role:   "performer",
			requestBody: discountdto.CreateDiscountRequest{
				Type:      "percentage",
				Value:     10,
				ValidFrom: time.Now(),
				ValidTo:   time.Now().Add(30 * 24 * time.Hour),
				MaxUses:   100,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Success - Create fixed amount discount",
			userID: TestPerformerID,
			role:   "performer",
			requestBody: discountdto.CreateDiscountRequest{
				Type:      "fixed_amount",
				Value:     100,
				ValidFrom: time.Now(),
				ValidTo:   time.Now().Add(30 * 24 * time.Hour),
				MaxUses:   50,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Bad Request - Invalid time range",
			userID: TestPerformerID,
			role:   "performer",
			requestBody: discountdto.CreateDiscountRequest{
				Type:      "percentage",
				Value:     10,
				ValidFrom: time.Now().Add(30 * 24 * time.Hour),
				ValidTo:   time.Now(),
				MaxUses:   100,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Forbidden - Client creates discount",
			userID: TestClientID,
			role:   "client",
			requestBody: discountdto.CreateDiscountRequest{
				Type:      "percentage",
				Value:     10,
				ValidFrom: time.Now(),
				ValidTo:   time.Now().Add(30 * 24 * time.Hour),
				MaxUses:   100,
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := GenerateTestToken(tt.userID, tt.role)
			body, _ := json.Marshal(tt.requestBody)
			url := fmt.Sprintf("/api/v1/services/%s/discounts", serviceID.String())
			req := httptest.NewRequest("POST", url, bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			res := httptest.NewRecorder()

			testRouter.ServeHTTP(res, req)

			assert.Equal(t, tt.expectedStatus, res.Code)

			if tt.expectedStatus == http.StatusOK {
				var response discountdto.DiscountResponse
				err := json.NewDecoder(res.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody.Type, response.Type)
				assert.Equal(t, tt.requestBody.Value, response.Value)
			}
		})
	}
}

func TestGetDiscount(t *testing.T) {
	SkipIfNoDocker(t)

	token := GenerateTestToken(TestClientID, "client")
	req := httptest.NewRequest("GET", "/api/v1/discounts/"+TestDiscountID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	testRouter.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)

	var response discountdto.DiscountResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, TestDiscountID.String(), response.ID)
	assert.Equal(t, TestServiceID.String(), response.ServiceID)
}

func TestUpdateDiscount(t *testing.T) {
	SkipIfNoDocker(t)
	defer CleanupTestData(t)

	// Create discount for update
	discountID := uuid.New()
	_, err := testDB.Exec(`
		INSERT INTO discounts (id, service_id, type, value, max_uses, valid_from, valid_to)
		VALUES ($1, $2, 'percentage', 10, 50, NOW(), NOW() + INTERVAL '30 days')`,
		discountID, TestServiceID)
	require.NoError(t, err)

	newMaxUses := 200
	newValidTo := time.Now().Add(60 * 24 * time.Hour)
	requestBody := discountdto.PatchDiscountRequest{
		MaxUses: &newMaxUses,
		ValidTo: &newValidTo,
	}

	token := GenerateTestToken(TestPerformerID, "performer")
	body, _ := json.Marshal(requestBody)
	url := fmt.Sprintf("/api/v1/services/%s/discounts/%s", TestServiceID.String(), discountID.String())
	req := httptest.NewRequest("PATCH", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	testRouter.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)

	// Verify update
	var maxUses int32
	err = testDB.QueryRow("SELECT max_uses FROM discounts WHERE id = $1", discountID).Scan(&maxUses)
	assert.NoError(t, err)
	assert.Equal(t, int32(newMaxUses), maxUses)
}

func TestDeleteDiscount(t *testing.T) {
	SkipIfNoDocker(t)
	defer CleanupTestData(t)

	// Create discount for deletion
	discountID := uuid.New()
	_, err := testDB.Exec(`
		INSERT INTO discounts (id, service_id, type, value, max_uses, valid_from, valid_to)
		VALUES ($1, $2, 'percentage', 10, 50, NOW(), NOW() + INTERVAL '30 days')`,
		discountID, TestServiceID)
	require.NoError(t, err)

	token := GenerateTestToken(TestPerformerID, "performer")
	url := fmt.Sprintf("/api/v1/services/%s/discounts/%s", TestServiceID.String(), discountID.String())
	req := httptest.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	testRouter.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)

	// Verify deletion
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM discounts WHERE id = $1", discountID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// ========== REVIEW TESTS ==========

func TestCreateReview(t *testing.T) {
	SkipIfNoDocker(t)
	defer CleanupTestData(t)

	// Create booking for review
	bookingID := uuid.New()
	_, err := testDB.Exec(`
		INSERT INTO bookings (id, client_id, service_id, base_price, final_price, booking_time, status)
		VALUES ($1, $2, $3, 1000, 1000, $4, 'completed')`,
		bookingID, TestClientID, TestServiceID, time.Now().Add(-24*time.Hour))
	require.NoError(t, err)

	tests := []struct {
		name           string
		userID         uuid.UUID
		role           string
		requestBody    reviewdto.CreateReviewRequest
		expectedStatus int
	}{
		{
			name:   "Success - Create review",
			userID: TestClientID,
			role:   "client",
			requestBody: reviewdto.CreateReviewRequest{
				BookingID: bookingID,
				Rating:    5,
				Comment:   "Excellent service!",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Bad Request - Invalid rating (too high)",
			userID: TestClientID,
			role:   "client",
			requestBody: reviewdto.CreateReviewRequest{
				BookingID: bookingID,
				Rating:    6,
				Comment:   "Invalid rating",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Bad Request - Invalid rating (too low)",
			userID: TestClientID,
			role:   "client",
			requestBody: reviewdto.CreateReviewRequest{
				BookingID: bookingID,
				Rating:    0,
				Comment:   "Invalid rating",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Forbidden - Wrong user creates review",
			userID: TestPerformerID,
			role:   "performer",
			requestBody: reviewdto.CreateReviewRequest{
				BookingID: bookingID,
				Rating:    4,
				Comment:   "Wrong user",
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := GenerateTestToken(tt.userID, tt.role)
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/reviews", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			res := httptest.NewRecorder()

			testRouter.ServeHTTP(res, req)

			assert.Equal(t, tt.expectedStatus, res.Code)
		})
	}
}

func TestGetReviewByID(t *testing.T) {
	SkipIfNoDocker(t)

	token := GenerateTestToken(TestClientID, "client")
	req := httptest.NewRequest("GET", "/api/v1/reviews/"+TestReviewID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	testRouter.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)

	var response reviewdto.ReviewResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, TestReviewID.String(), response.ID)
	assert.Equal(t, int32(5), response.Rating)
}

func TestGetReviewByBookingID(t *testing.T) {
	SkipIfNoDocker(t)

	token := GenerateTestToken(TestClientID, "client")
	req := httptest.NewRequest("GET", "/api/v1/booking/"+TestBookingID.String()+"/reviews", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	testRouter.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)

	var response reviewdto.ReviewResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, TestBookingID.String(), response.BookingID)
}

func TestGetReviewsByServiceID(t *testing.T) {
	SkipIfNoDocker(t)

	tests := []struct {
		name           string
		page           string
		limit          string
		expectedStatus int
	}{
		{
			name:           "Success - Get reviews with pagination",
			page:           "1",
			limit:          "10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Success - Get reviews with limit",
			page:           "1",
			limit:          "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Bad Request - Invalid page",
			page:           "invalid",
			limit:          "10",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := GenerateTestToken(TestClientID, "client")
			url := fmt.Sprintf("/api/v1/service/%s/reviews?page=%s&limit=%s", TestServiceID.String(), tt.page, tt.limit)
			req := httptest.NewRequest("GET", url, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			res := httptest.NewRecorder()

			testRouter.ServeHTTP(res, req)

			assert.Equal(t, tt.expectedStatus, res.Code)

			if tt.expectedStatus == http.StatusOK {
				var reviews []reviewdto.ReviewResponse
				err := json.NewDecoder(res.Body).Decode(&reviews)
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateReview(t *testing.T) {
	SkipIfNoDocker(t)
	defer CleanupTestData(t)

	// Create review for update
	reviewID := uuid.New()
	_, err := testDB.Exec(`
		INSERT INTO reviews (id, booking_id, rating, comment)
		VALUES ($1, $2, 4, 'Original comment')`,
		reviewID, TestBookingID)
	require.NoError(t, err)

	newRating := int32(5)
	newComment := "Updated comment"
	requestBody := reviewdto.PatchReviewRequest{
		Rating:  &newRating,
		Comment: &newComment,
	}

	token := GenerateTestToken(TestClientID, "client")
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PATCH", "/api/v1/reviews/"+reviewID.String(), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	testRouter.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)

	// Verify update
	var rating int32
	var comment sql.NullString
	err = testDB.QueryRow("SELECT rating, comment FROM reviews WHERE id = $1", reviewID).Scan(&rating, &comment)
	assert.NoError(t, err)
	assert.Equal(t, newRating, rating)
	assert.True(t, comment.Valid)
	assert.Equal(t, newComment, comment.String)
}

func TestDeleteReview(t *testing.T) {
	SkipIfNoDocker(t)
	defer CleanupTestData(t)

	// Create review for deletion
	reviewID := uuid.New()
	_, err := testDB.Exec(`
		INSERT INTO reviews (id, booking_id, rating, comment)
		VALUES ($1, $2, 4, 'To be deleted')`,
		reviewID, TestBookingID)
	require.NoError(t, err)

	token := GenerateTestToken(TestClientID, "client")
	req := httptest.NewRequest("DELETE", "/api/v1/reviews/"+reviewID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	testRouter.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)

	// Verify deletion
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM reviews WHERE id = $1", reviewID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// ========== ADMIN TESTS ==========

func TestAdminUpdateVerificationStatus(t *testing.T) {
	SkipIfNoDocker(t)
	defer CleanupTestData(t)

	tests := []struct {
		name           string
		userID         uuid.UUID
		role           string
		targetUserID   uuid.UUID
		status         string
		expectedStatus int
	}{
		{
			name:           "Success - Admin verifies performer",
			userID:         TestAdminID,
			role:           "admin",
			targetUserID:   TestUnverifiedID,
			status:         "verified",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Success - Admin rejects performer",
			userID:         TestAdminID,
			role:           "admin",
			targetUserID:   TestUnverifiedID,
			status:         "rejected",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Forbidden - Performer tries to verify",
			userID:         TestPerformerID,
			role:           "performer",
			targetUserID:   TestUnverifiedID,
			status:         "verified",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := profiledto.UpdateVerificationStatusRequest{
				UserID:             tt.targetUserID,
				VerificationStatus: tt.status,
			}
			body, _ := json.Marshal(requestBody)
			token := GenerateTestToken(tt.userID, tt.role)
			req := httptest.NewRequest("PATCH", "/api/v1/profile/verification_status", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			res := httptest.NewRecorder()

			testRouter.ServeHTTP(res, req)

			assert.Equal(t, tt.expectedStatus, res.Code)

			if tt.expectedStatus == http.StatusNoContent {
				// Verify status was updated
				var status string
				err := testDB.QueryRow("SELECT verification_status FROM users WHERE id = $1", tt.targetUserID).Scan(&status)
				assert.NoError(t, err)
				assert.Equal(t, tt.status, status)
			}
		})
	}
}

// ========== UNAUTHORIZED ACCESS TESTS ==========

func TestUnauthorizedAccess(t *testing.T) {
	SkipIfNoDocker(t)

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/profile"},
		{"PATCH", "/api/v1/profile"},
		{"DELETE", "/api/v1/profile"},
		{"GET", "/api/v1/services"},
		{"POST", "/api/v1/services"},
		{"GET", "/api/v1/services/" + TestServiceID.String()},
		{"POST", "/api/v1/reviews"},
		{"GET", "/api/v1/reviews/" + TestReviewID.String()},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, nil)
			res := httptest.NewRecorder()

			testRouter.ServeHTTP(res, req)

			assert.Equal(t, http.StatusUnauthorized, res.Code)
		})
	}
}

// ========== MIDDLEWARE TESTS ==========

func TestRoleMiddleware(t *testing.T) {
	SkipIfNoDocker(t)

	// Endpoint that requires admin/performer role
	endpoint := "/api/v1/services" // POST requires admin/performer

	tests := []struct {
		name           string
		role           string
		expectedStatus int
	}{
		{"Admin role allowed", "admin", http.StatusOK},
		{"Performer role allowed", "performer", http.StatusOK},
		{"Client role forbidden", "client", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var userID uuid.UUID
			switch tt.role {
			case "admin":
				userID = TestAdminID
			case "performer":
				userID = TestPerformerID
			default:
				userID = TestClientID
			}

			token := GenerateTestToken(userID, tt.role)
			requestBody := servicedto.CreateServiceRequest{
				Title:           "Test",
				Description:     "Test",
				Price:           1000,
				DurationMinutes: 60,
			}
			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", endpoint, bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			res := httptest.NewRecorder()

			testRouter.ServeHTTP(res, req)

			assert.Equal(t, tt.expectedStatus, res.Code)
		})
	}
}

// ========== EDGE CASES TESTS ==========

func TestNonExistentResource(t *testing.T) {
	SkipIfNoDocker(t)

	nonExistentID := uuid.New()
	token := GenerateTestToken(TestClientID, "client")

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/services/" + nonExistentID.String()},
		{"GET", "/api/v1/discounts/" + nonExistentID.String()},
		{"GET", "/api/v1/reviews/" + nonExistentID.String()},
		{"GET", "/api/v1/booking/" + nonExistentID.String() + "/reviews"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			res := httptest.NewRecorder()

			testRouter.ServeHTTP(res, req)

			assert.Equal(t, http.StatusInternalServerError, res.Code)
		})
	}
}
