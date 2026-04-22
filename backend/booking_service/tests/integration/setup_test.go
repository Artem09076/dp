//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Artem09076/dp/backend/booking_service/internal/application/booking"
	"github.com/Artem09076/dp/backend/booking_service/internal/config"
	"github.com/Artem09076/dp/backend/booking_service/internal/lib/jwt"
	bookinghandlers "github.com/Artem09076/dp/backend/booking_service/internal/presentation/booking/handlers"
	bookingmiddleware "github.com/Artem09076/dp/backend/booking_service/internal/presentation/middleware"
	sqlc "github.com/Artem09076/dp/backend/booking_service/internal/storage/db"
	redisstorage "github.com/Artem09076/dp/backend/booking_service/internal/storage/redis"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

var (
	testDB          *sql.DB
	testRedisClient *redis.Client
	testRabbitConn  *amqp.Connection
	testRabbitCh    *amqp.Channel
	testRouter      *chi.Mux
	testQueries     *sqlc.Queries
	testLogger      *slog.Logger
	testConfig      *config.Config

	TestServiceID   = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	TestDiscountID  = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	TestPerformerID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	TestClientID    = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	TestAdminID     = uuid.MustParse("55555555-5555-5555-5555-555555555555")

	jwtSecret = "test-secret-key-for-integration-tests"
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

	postgresPort := postgresResource.GetPort("5432/tcp")
	testDB, err = sql.Open("postgres", "postgres://postgres:test@localhost:"+postgresPort+"/testdb?sslmode=disable")
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

	if err := runMigrations(testDB); err != nil {
		testLogger.Error("Could not run migrations", "error", err)
		os.Exit(1)
	}

	if err := seedTestData(testDB); err != nil {
		testLogger.Error("Could not seed test data", "error", err)
		os.Exit(1)
	}

	var serviceCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM services").Scan(&serviceCount)
	if err != nil {
		testLogger.Error("Could not verify services", "error", err)
		os.Exit(1)
	}
	testLogger.Info("Services count", "count", serviceCount)

	var userCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		testLogger.Error("Could not verify users", "error", err)
		os.Exit(1)
	}
	testLogger.Info("Users count", "count", userCount)

	redisPort := redisResource.GetPort("6379/tcp")
	testRedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:" + redisPort,
		DB:   0,
	})

	err = pool.Retry(func() error {
		return testRedisClient.Ping(context.Background()).Err()
	})
	if err != nil {
		testLogger.Error("Could not connect to redis", "error", err)
		os.Exit(1)
	}
	testRouter = setupTestRouter(testDB, testRedisClient)
	testQueries = sqlc.New(testDB)

	code := m.Run()

	if testRabbitCh != nil {
		testRabbitCh.Close()
	}
	if testRabbitConn != nil {
		testRabbitConn.Close()
	}
	if testRedisClient != nil {
		testRedisClient.Close()
	}
	if testDB != nil {
		testDB.Close()
	}

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
	enumTypes := []string{
		`DROP TYPE IF EXISTS user_role CASCADE`,
		`DROP TYPE IF EXISTS business_type CASCADE`,
		`DROP TYPE IF EXISTS verification_status CASCADE`,
		`DROP TYPE IF EXISTS booking_status CASCADE`,
		`DROP TYPE IF EXISTS discoutn_type CASCADE`,

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

	migrations := []string{
		`DROP TABLE IF EXISTS reviews CASCADE`,
		`DROP TABLE IF EXISTS bookings CASCADE`,
		`DROP TABLE IF EXISTS discounts CASCADE`,
		`DROP TABLE IF EXISTS services CASCADE`,
		`DROP TABLE IF EXISTS users CASCADE`,

		`CREATE TABLE IF NOT EXISTS users (
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

		`CREATE TABLE IF NOT EXISTS services (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			performer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			price BIGINT NOT NULL,
			duration_minutes INT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS discounts (
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

		`CREATE TABLE IF NOT EXISTS bookings (
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

		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_services_performer ON services(performer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bookings_client ON bookings(client_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bookings_service ON bookings(service_id)`,

		`CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
		   NEW.updated_at = NOW();
		   RETURN NEW;
		END;
		$$ language 'plpgsql'`,

		`CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,

		`CREATE TRIGGER update_services_updated_at BEFORE UPDATE ON services
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,

		`CREATE TRIGGER update_bookings_updated_at BEFORE UPDATE ON bookings
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %s, error: %w", i, migration, err)
		}
	}
	return nil
}

func seedTestData(db *sql.DB) error {
	passwordHash := []byte("$2a$10$test_hash_for_integration_tests")

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", TestClientID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check users: %w", err)
	}

	if count == 0 {
		_, err = db.Exec(`
			INSERT INTO users (id, email, name, password_hash, role) VALUES 
			($1, 'client@test.com', 'Test Client', $4, 'client'),
			($2, 'performer@test.com', 'Test Performer', $4, 'performer'),
			($3, 'admin@test.com', 'Test Admin', $4, 'admin')`,
			TestClientID, TestPerformerID, TestAdminID, passwordHash)
		if err != nil {
			return fmt.Errorf("failed to insert users: %w", err)
		}
	}

	err = db.QueryRow("SELECT COUNT(*) FROM services WHERE id = $1", TestServiceID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check services: %w", err)
	}

	if count == 0 {
		_, err = db.Exec(`
			INSERT INTO services (id, title, description, price, duration_minutes, performer_id) VALUES 
			($1, 'Test Service', 'Test Description', 1000, 60, $2)`,
			TestServiceID, TestPerformerID)
		if err != nil {
			return fmt.Errorf("failed to insert services: %w", err)
		}
	}

	err = db.QueryRow("SELECT COUNT(*) FROM discounts WHERE id = $1", TestDiscountID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check discounts: %w", err)
	}

	if count == 0 {
		_, err = db.Exec(`
			INSERT INTO discounts (id, service_id, type, value, max_uses, valid_from, valid_to) VALUES 
			($1, $2, 'percentage', 20, 100, NOW() - INTERVAL '1 day', NOW() + INTERVAL '30 days')`,
			TestDiscountID, TestServiceID)
		if err != nil {
			return fmt.Errorf("failed to insert discounts: %w", err)
		}
	}

	testLogger.Info("Test data seeded successfully",
		"service_id", TestServiceID,
		"performer_id", TestPerformerID,
		"client_id", TestClientID)

	return nil
}

type MockPublisher struct {
	PublishedMessages []interface{}
}

func (m *MockPublisher) Publish(queue string, msg interface{}) error {
	m.PublishedMessages = append(m.PublishedMessages, msg)
	return nil
}

func setupTestRouter(db *sql.DB, redisClient *redis.Client) *chi.Mux {
	queries := sqlc.New(db)

	redisWrapper, err := redisstorage.NewRedisClient(testLogger, redisClient.Options().Addr, "", 0)
	if err != nil {
		testLogger.Error("Failed to create redis client", "error", err)
		return nil
	}

	tokenValidator := jwt.NewValidator(jwtSecret, redisWrapper)

	mockPublisher := &MockPublisher{
		PublishedMessages: make([]interface{}, 0),
	}

	bookingService := booking.NewBookingService(queries, db, testLogger, mockPublisher)
	bookingHandlers := bookinghandlers.NewBookingHandler(bookingService, testLogger)

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)

	router.Group(func(r chi.Router) {
		r.Use(bookingmiddleware.NewJWTMiddleware(testLogger, tokenValidator))
		r.Post("/api/v1/bookings", bookingHandlers.CreateBooking())
		r.Patch("/api/v1/bookings/cancel/{id}", bookingHandlers.CancelBooking())
		r.Patch("/api/v1/bookings/submit/{id}", bookingHandlers.SubmitBooking())
		r.Patch("/api/v1/bookings/{id}", bookingHandlers.PatchBooking())
		r.Get("/api/v1/bookings/{id}", bookingHandlers.GetBooking())
		r.Get("/api/v1/bookings", bookingHandlers.GetBookings())
		r.Delete("/api/v1/bookings/{id}", bookingHandlers.DeleteBooking())
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
	_, err := testDB.Exec("DELETE FROM bookings")
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
