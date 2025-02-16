package main

import (
	"context"
	"fmt"
	"github.com/gratefultolord/merch-store/internal/config"
	"github.com/gratefultolord/merch-store/internal/handlers"
	mw "github.com/gratefultolord/merch-store/internal/middleware"
	"github.com/gratefultolord/merch-store/internal/repository"
	"github.com/gratefultolord/merch-store/internal/services"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	time.Sleep(5 * time.Second)
	fmt.Println(cfg.GetDatabaseURL())

	db, err := sqlx.Connect("postgres", cfg.GetDatabaseURL())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepo(db)
	itemRepo := repository.NewItemRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)

	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	coinService := services.NewCoinService(userRepo, transactionRepo, db)
	inventoryService := services.NewInventoryService(userRepo, itemRepo, transactionRepo, db)
	infoService := services.NewInfoService(userRepo, coinService)

	authHandler := handlers.NewAuthHandler(authService)
	sendCoinHandler := handlers.NewSendCoinHandler(coinService, userRepo)
	buyHandler := handlers.NewBuyHandler(inventoryService, userRepo, itemRepo)
	infoHandler := handlers.NewInfoHandler(infoService)

	e := echo.New()

	e.Use(middleware.Logger())

	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodPost,
			http.MethodDelete},
	}))

	e.POST("/api/auth", authHandler.Auth)

	authGroup := e.Group("")
	authMiddleware := mw.NewAuthMiddleware(userRepo, cfg.JWTSecret)
	authGroup.Use(authMiddleware)

	authGroup.POST("/api/sendCoin", sendCoinHandler.SendCoin)
	authGroup.POST("/api/buy/:item", buyHandler.Buy)
	authGroup.GET("/api/info", infoHandler.Info)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.ServerPort)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	log.Println("Server exited properly")
}
