// main.go

package main

import (
	"context"
	"log"
	"productproject/internal/config"
	"productproject/internal/handlers"

	product "productproject/internal/product"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := product.NewPostgresDatabase(cfg.GetConnectionString())
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
	}
	if db != nil {
		defer db.Close()
	}

	store := product.NewStore(db)
	h := handlers.NewProductHandlers(store)

	go func() {
		for {
			time.Sleep(10 * time.Second)
			if err := db.Ping(); err != nil {
				log.Printf("Database connection lost: %v", err)
				// พยายามเชื่อมต่อใหม่
				if reconnErr := db.Reconnect(cfg.GetConnectionString()); reconnErr != nil {
					log.Printf("Failed to reconnect: %v", reconnErr)
				} else {
					log.Printf("Successfully reconnected to the database")
				}
			}
		}
	}()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// กำหนดค่า CORS
	configCors := cors.Config{
		AllowOrigins:     []string{"*"}, // "*" ยอมรับทุกโดเมน
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(configCors))

	r.Use(TimeoutMiddleware(5 * time.Second))

	r.GET("/health", h.HealthCheck)

	// API v1
	v1 := r.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.GET("/:id", h.GetProduct)
			products.GET("/allproducts", h.AllProducts)
			products.GET("/recommend", h.GetProductRecommend)
			products.GET("/new", h.GetNewProduct)
			products.GET("/search", h.SearchProduct)
			products.GET("/category/:category", h.GetProductByCategory)
		}
		seller := v1.Group("/seller")
		{
			seller.GET("/:id", h.GetSeller)
		}
		cart := v1.Group("/cart")
		{
			cart.GET("/allcart", h.GetAllCartItems)
			cart.POST("/addcart", h.AddToCart)
			cart.PUT("/updatecart", h.UpdateCartItemQuantity)
			cart.DELETE("/deletecart", h.DeleteCartItem)
		}
		// User
		users := v1.Group("/users")
		{
			// ใช้ UserHandlers สำหรับเส้นทางที่เกี่ยวข้องกับผู้ใช้
			userHandlers := handlers.NewUserHandlers(store) // สร้าง instance ของ UserHandlers

			// เส้นทาง "/users/me" สำหรับดึงข้อมูลของผู้ใช้ที่ล็อกอินอยู่
			users.GET("/me", userHandlers.GetUserProfile)

			// เส้นทาง "/users/:user_id" สำหรับดึงข้อมูลของผู้ใช้ที่ระบุ
			users.GET("/:user_id", userHandlers.GetUserProfile)
			users.PUT("/updateuser", h.UpdateUserContactHandler)
		}

		order := v1.Group("/order")
		{
			order.POST("/create", h.CreateOrder)
			order.GET("/allorder", h.GetOrders)
			order.GET("/:status", h.GetOrdersSort)
			order.PUT("/update", h.UpdateCartItemStatusHandler)
		}
	}

	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Printf("Failed to run server: %v", err)
	}
}
