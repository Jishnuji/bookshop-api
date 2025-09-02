package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	_ "sync"
	"syscall"
	"time"
	"toptal/internal/app/config"
	"toptal/internal/app/repository/pgrepo"
	"toptal/internal/app/services"
	"toptal/internal/app/transport/httpserver"

	"toptal/internal/pkg/pg"

	"database/sql"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func run() error {
	cfg, err := config.Read()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	log.Printf("Config loaded: %v", cfg)

	pgDB, err := pg.Dial(cfg.DSN)
	if err != nil {
		return fmt.Errorf("pg.Dial failed: %w", err)
	}

	if pgDB != nil {
		log.Printf("Run PostgreSQL mirgations")
		err := migratePgData(cfg.MigrationsPath, cfg.DSN)
		if err != nil {
			return fmt.Errorf("migratePgData failed: %w", err)
		}

	}

	// create repositories
	userRepo := pgrepo.NewUserRepo(pgDB)
	bookRepo := pgrepo.NewBookRepository(pgDB)
	categoryRepo := pgrepo.NewCategoryRepository(pgDB)
	cartRepo := pgrepo.NewCartRepository(pgDB)

	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo)
	bookService := services.NewBookService(bookRepo)
	categoryService := services.NewCategoryService(categoryRepo)
	cartService := services.NewCartService(cartRepo)

	// create http server
	httpServer := httpserver.NewHttpServer(userService, authService, bookService, cartService, categoryService)

	// create router
	router := chi.NewRouter()

	// add base middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.Heartbeat("/health"))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Book Shop API v0.1"))
	})

	// Public routes (no auth needed)
	router.Group(func(r chi.Router) {
		// Auth
		r.Post("/signup", httpServer.SignUp)
		r.Post("/signin", httpServer.SignIn)

		// Books
		r.Get("/books", httpServer.GetBooks)
		router.Get("/book/{book_id}", httpServer.GetBook)

		// Categories
		r.Get("/categories", httpServer.GetCategories)
		r.Get("/category/{category_id}", httpServer.GetCategory)
	})

	// Protected routes (auth needed)
	router.Group(func(r chi.Router) {
		r.Use(httpServer.CheckAuthorizedUser)

		//Cart
		r.Post("/cart", httpServer.UpdateCart)
		r.Post("/checkout", httpServer.Checkout)
	})

	// Admin routes (admin auth needed)
	router.Group(func(r chi.Router) {
		r.Use(httpServer.CheckAdmin)

		// Books
		r.Post("/book", httpServer.CreateBook)
		r.Patch("/book/{book_id}", httpServer.UpdateBook)
		r.Delete("/book/{book_id}", httpServer.DeleteBook)

		// Categories
		r.Post("/category", httpServer.CreateCategory)
		r.Patch("/category/{category_id}", httpServer.UpdateCategory)
		r.Delete("/category/{category_id}", httpServer.DeleteCategory)
	})

	// Clean expired carts every minute
	ctx, cleanupCancel := context.WithCancel(context.Background())
	cleanupFinished := make(chan struct{})
	go func() {
		defer close(cleanupFinished)
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Println("Cleaning expired carts")
				err := cartRepo.CleanExpiredCarts(ctx, 30*time.Minute)
				if err != nil {
					log.Printf("cartRepo.CleanExpiredCarts failed: %v", err)
				}
			case <-ctx.Done():
				log.Println("Cart cleanup goroutine stopped")
				return
			}
		}
	}()

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	// listen to OS signals and gracefully shutdown HTTP server
	serverStopped := make(chan struct{})
	go func() {
		defer close(serverStopped)
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		cleanupCancel() // stop cart cleanup goroutine

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}
	}()

	log.Printf("Starting HTTP server on %s", cfg.HTTPAddr)

	// start HTTP server
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe Error: %v", err)
	}

	//Wait for goroutines to finish
	<-serverStopped
	<-cleanupFinished

	log.Printf("Have a nice day!")

	return nil
}

// migratePgData runs Postgres migrations
func migratePgData(path string, dsn string) error {
	if dsn == "" {
		return errors.New("dsn is empty")
	}
	if path == "" {
		return errors.New("migrations path is empty")
	}

	// Remove "file://" prefix if present
	if strings.HasPrefix(path, "file://") {
		path = strings.TrimPrefix(path, "file://")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	defer db.Close()

	goose.SetTableName("goose_migrations")

	if err := goose.Up(db, path); err != nil {
		return fmt.Errorf("failed to apply migrations from %s: %w", path, err)
	}

	return nil
}
