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
	"sync"
	_ "sync"
	"syscall"
	"time"
	"toptal/internal/app/config"
	"toptal/internal/app/repository/pgrepo"
	"toptal/internal/app/services"
	"toptal/internal/app/transport/grpcserver"
	"toptal/internal/app/transport/httpserver"
	authv1 "toptal/proto/v1/auth"
	bookv1 "toptal/proto/v1/book"
	cartv1 "toptal/proto/v1/cart"
	categoryv1 "toptal/proto/v1/category"

	"toptal/internal/pkg/pg"

	"database/sql"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

	// create grpc server
	grpcServer := grpcserver.NewGrpcServer(userService, authService, bookService, cartService, categoryService)

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

	err = addGrpcEndpoints(router, cfg.GRPCAddr, httpServer)
	if err != nil {
		return fmt.Errorf("failed to add gRPC gateway routes: %w", err)
	}

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

		grpcServer.Stop()
	}()

	var wg sync.WaitGroup

	// start HTTP server
	wg.Go(func() {
		log.Printf("Starting HTTP server on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe Error: %v", err)
		}
		log.Printf("HTTP server stopped")
	})

	// start GRPC server
	wg.Go(func() {
		log.Printf("Starting GRPC server on %s", cfg.GRPCAddr)
		if err := grpcServer.Start(cfg.GRPCAddr); err != nil {
			log.Fatalf("GRPC server ListenAndServe Error: %v", err)
		}
		log.Printf("gRPC server stopped")
	})

	//Wait for goroutines to finish
	<-serverStopped
	<-cleanupFinished
	wg.Wait()

	log.Printf("Have a nice day!")

	return nil
}

func addGrpcEndpoints(router *chi.Mux, addr string, httpServer *httpserver.HttpServer) error {
	ctx := context.Background()
	// TODO: add interceptor for transport validation and for RPC calls
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	gwMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			switch strings.ToLower(key) {
			case "authorization":
				return key, true
			case "user-id", "user-email", "user-admin":
				return key, true
			default:
				return runtime.DefaultHeaderMatcher(key)
			}
		}))

	err := authv1.RegisterAuthServiceHandlerFromEndpoint(ctx, gwMux, addr, opts)
	if err != nil {
		return fmt.Errorf("failed to register auth service handler: %w", err)
	}

	err = bookv1.RegisterBookServiceHandlerFromEndpoint(ctx, gwMux, addr, opts)
	if err != nil {
		return fmt.Errorf("failed to register book service handler: %w", err)
	}

	err = categoryv1.RegisterCategoryServiceHandlerFromEndpoint(ctx, gwMux, addr, opts)
	if err != nil {
		return fmt.Errorf("failed to register category service handler: %w", err)
	}

	err = categoryv1.RegisterCategoryServiceHandlerFromEndpoint(ctx, gwMux, addr, opts)
	if err != nil {
		return fmt.Errorf("failed to register category service handler: %w", err)
	}

	err = cartv1.RegisterCartServiceHandlerFromEndpoint(ctx, gwMux, addr, opts)
	if err != nil {
		return fmt.Errorf("failed to register cart service handler: %w", err)
	}

	gwRouter := chi.NewRouter()
	gwRouter.Mount("/", gwMux)
	router.Mount("/v1", gwRouter)

	// Protected routes (auth needed)
	router.Group(func(r chi.Router) {
		r.Use(httpServer.CheckAuthorizedUser)

		//Cart
		r.Post("/v1/cart", gwMux.ServeHTTP)
		r.Post("/v1/checkout", gwMux.ServeHTTP)
	})

	// Admin routes (admin auth needed)
	router.Group(func(r chi.Router) {
		r.Use(httpServer.CheckAdmin)

		// Books
		r.Post("/v1/book", gwMux.ServeHTTP)
		r.Patch("/v1/book/{book_id}", gwMux.ServeHTTP)
		r.Delete("/v1/book/{book_id}", gwMux.ServeHTTP)

		// Categories
		r.Post("/v1/category", gwMux.ServeHTTP)
		r.Patch("/v1/category/{category_id}", gwMux.ServeHTTP)
		r.Delete("/v1/category/{category_id}", gwMux.ServeHTTP)
	})
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
