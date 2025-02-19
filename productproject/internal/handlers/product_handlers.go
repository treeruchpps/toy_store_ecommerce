// product_handlers.go
package handlers

import (
	"fmt"
	"log"
	"net/http"
	product "productproject/internal/product"
	user "productproject/internal/product"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandlers struct {
	store *product.Store
}

type UserHandlers struct {
	store *user.Store
}

func NewProductHandlers(store *product.Store) *ProductHandlers {
	return &ProductHandlers{store: store}
}

func NewUserHandlers(store *user.Store) *UserHandlers {
	return &UserHandlers{store: store}
}

func convertTimesToUserTimezone(product *product.ProductItem, loc *time.Location) {
	product.CreatedAt = product.CreatedAt.In(loc)
	product.UpdatedAt = product.UpdatedAt.In(loc)
	product.Inventory.UpdatedAt = product.Inventory.UpdatedAt.In(loc)
}

func (h *ProductHandlers) GetProduct(c *gin.Context) {
	id := c.Param("id")

	product, err := h.store.GetProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	userTimezone := "Asia/Bangkok"
	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		log.Fatal("ไม่สามารถโหลด timezone ได้:", err)
	}

	convertTimesToUserTimezone(&product, loc)

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandlers) AllProducts(c *gin.Context) {
	allProducts, err := h.store.AllProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลสินค้าล่าสุดได้"})
		return
	}

	userTimezone := "Asia/Bangkok"
	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		log.Fatal("ไม่สามารถโหลด timezone ได้:", err)
	}

	for i := range allProducts {
		convertTimesToUserTimezone(&allProducts[i], loc)
	}

	c.JSON(http.StatusOK, allProducts)
}

func (h ProductHandlers) GetSeller(c *gin.Context) {
	// รับค่า seller_id จาก URL parameter
	id := c.Param("id")

	// เรียกใช้ method GetSeller จาก store เพื่อดึงข้อมูลร้านค้า
	seller, err := h.store.GetSeller(c.Request.Context(), id)
	if err != nil {
		// หากเกิดข้อผิดพลาด (ไม่พบร้านค้า)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// ส่งข้อมูลร้านค้ากลับในรูปแบบ JSON
	c.JSON(http.StatusOK, seller)
}

func (h *ProductHandlers) GetProductRecommend(c *gin.Context) {
	// ดึงรายการสินค้าที่แนะนำจาก store
	recommendedProducts, err := h.store.GetProductRecommend(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลสินค้าแนะนำได้"})
		return
	}

	// ตั้งค่า timezone
	userTimezone := "Asia/Bangkok"
	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		log.Fatal("ไม่สามารถโหลด timezone ได้:", err)
	}

	// เปลี่ยนเวลาเป็น timezone ของผู้ใช้
	for i := range recommendedProducts {
		convertTimesToUserTimezone(&recommendedProducts[i], loc)
	}

	// ส่งข้อมูลสินค้าแนะนำเป็น JSON
	c.JSON(http.StatusOK, recommendedProducts)
}

func (h *ProductHandlers) GetNewProduct(c *gin.Context) {
	newProducts, err := h.store.GetNewProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลสินค้าล่าสุดได้"})
		return
	}

	userTimezone := "Asia/Bangkok"
	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		log.Fatal("ไม่สามารถโหลด timezone ได้:", err)
	}

	for i := range newProducts {
		convertTimesToUserTimezone(&newProducts[i], loc)
	}

	c.JSON(http.StatusOK, newProducts)
}

func (h *ProductHandlers) SearchProduct(c *gin.Context) {
	// รับค่า query จาก URL
	query := c.DefaultQuery("query", "") // ใช้ DefaultQuery หากไม่มี query จะส่งค่าเริ่มต้นเป็น ""

	// ตรวจสอบว่าผู้ใช้กรอกคำค้นหามาหรือไม่
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณากรอกคำค้นหา"})
		return
	}

	// Log ค่าที่ได้รับจาก query เพื่อใช้ในการดีบัก
	log.Printf("Searching products with query: %s", query)

	// ค้นหาผลิตภัณฑ์จาก store
	products, err := h.store.SearchProducts(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถค้นหาผลิตภัณฑ์ได้"})
		return
	}

	// ตรวจสอบว่ามีสินค้าตรงกับคำค้นหาหรือไม่
	if len(products) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "ไม่พบสินค้าตามคำค้นหาของคุณ"})
		return
	}

	// ตั้งค่า timezone ของผู้ใช้
	userTimezone := "Asia/Bangkok"
	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		log.Fatal("ไม่สามารถโหลด timezone ได้:", err)
	}

	// แปลงวันที่ให้เป็น timezone ของผู้ใช้
	for i := range products {
		convertTimesToUserTimezone(&products[i], loc)
	}

	// ส่งข้อมูลสินค้าเป็น JSON
	c.JSON(http.StatusOK, products)
}

func (h *ProductHandlers) GetProductByCategory(c *gin.Context) {
	// รับค่า category จาก URL parameter
	category := c.Param("category")

	// ตรวจสอบว่าหมวดหมู่ที่ส่งมาไม่ว่างเปล่า
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุหมวดหมู่สินค้า"})
		return
	}

	// ดึงรายการสินค้าตามหมวดหมู่จาก store
	products, err := h.store.GetProductByCategory(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลสินค้าตามหมวดหมู่ได้"})
		return
	}

	// ตรวจสอบว่ามีสินค้าตามหมวดหมู่ที่ค้นหาหรือไม่
	if len(products) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "ไม่พบสินค้าตามหมวดหมู่ที่ระบุ"})
		return
	}

	// ตั้งค่า timezone ของผู้ใช้
	userTimezone := "Asia/Bangkok"
	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		log.Fatal("ไม่สามารถโหลด timezone ได้:", err)
	}

	// แปลงวันที่ให้เป็น timezone ของผู้ใช้
	for i := range products {
		convertTimesToUserTimezone(&products[i], loc)
	}

	// ส่งข้อมูลสินค้าเป็น JSON
	c.JSON(http.StatusOK, products)
}

func (h *ProductHandlers) AddToCart(c *gin.Context) {
	var input struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}

	// ตรวจสอบข้อมูล JSON ที่รับเข้ามา
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	// ตรวจสอบค่า Quantity ว่ามากกว่า 0 หรือไม่
	if input.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be greater than zero"})
		return
	}

	// เรียกใช้ AddToCart พร้อม CartID
	err := h.store.AddToCart(c.Request.Context(), input.ProductID, input.Quantity)
	if err != nil {
		log.Printf("Error adding product to cart: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product added to cart successfully"})
}

func (h *ProductHandlers) GetAllCartItems(c *gin.Context) {
	// เรียกใช้ method GetAllCartItems จาก store เพื่อดึงข้อมูลสินค้าจากทุกตะกร้า
	cartItems, err := h.store.GetAllCartItems(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่งข้อมูลสินค้ากลับในรูปแบบ JSON
	c.JSON(http.StatusOK, cartItems)
}

func (h *ProductHandlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func (h *UserHandlers) GetUserProfile(c *gin.Context) {
	// ดึง userID จาก URL parameters
	userID := c.Param("user_id")

	// ตรวจสอบว่า userID มีค่าหรือไม่
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// เรียกใช้ฟังก์ชัน GetUserByID จาก store เพื่อดึงข้อมูลผู้ใช้
	user, err := h.store.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		// ถ้าไม่พบผู้ใช้ หรือเกิดข้อผิดพลาดในการดึงข้อมูล
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// ส่งข้อมูลผู้ใช้กลับไปในรูป JSON
	c.JSON(http.StatusOK, gin.H{
		"user_id":             user.UserID,
		"display_name":        user.DisplayName,
		"address":             user.Address,
		"phone":               user.Phone,
		"email":               user.Email,
		"profile_picture_url": user.ProfilePictureURL,
		"status":              user.Status,
		"role":                user.Role,
		"created_at":          user.CreatedAt,
		"updated_at":          user.UpdatedAt,
	})
}

func (h *ProductHandlers) UpdateCartItemQuantity(c *gin.Context) {
	var input struct {
		CartItemID string  `json:"cart_item_id"`
		Quantity   int     `json:"quantity"`
		TotalPrice float64 `json:"total_price"`
	}

	// ตรวจสอบว่าได้รับข้อมูล JSON ที่ถูกต้องหรือไม่
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Invalid input format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	// เรียกใช้ฟังก์ชันจาก database layer เพื่ออัปเดตหรือลบรายการสินค้าในตะกร้า
	if input.Quantity == 0 {
		// ถ้า Quantity เป็น 0 ให้ลบรายการสินค้าออกจากตะกร้า
		err := h.store.DeleteCartItem(c.Request.Context(), input.CartItemID)
		if err != nil {
			log.Printf("Error deleting cart item: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Cart item deleted successfully"})
		return
	}

	// อัปเดตจำนวนสินค้าในตะกร้า
	err := h.store.UpdateCartItemQuantity(c.Request.Context(), input.CartItemID, input.Quantity)
	if err != nil {
		log.Printf("Error updating cart item quantity: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cart item quantity and total price updated successfully"})
}

// DeleteCartItem handler function
func (h *ProductHandlers) DeleteCartItem(c *gin.Context) {
	// สร้าง struct เพื่อรับข้อมูลจาก body
	var requestBody struct {
		CartItemID string `json:"cart_item_id"`
	}

	// อ่านข้อมูลจาก body ของคำขอ
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// ตรวจสอบว่าได้ระบุ cart_item_id หรือไม่
	if requestBody.CartItemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart item ID is required"})
		return
	}

	// เรียกฟังก์ชันใน database layer เพื่อลบสินค้าจากตะกร้า
	err := h.store.DeleteCartItem(c.Request.Context(), requestBody.CartItemID)
	if err != nil {
		log.Printf("Error deleting cart item: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart item"})
		return
	}

	// ส่งข้อความตอบกลับเมื่อลบสำเร็จ
	c.JSON(http.StatusOK, gin.H{"message": "Cart item deleted successfully"})
}

func (h *ProductHandlers) GetOrders(c *gin.Context) {
	// เรียกฟังก์ชันใน database layer เพื่อดึงข้อมูลคำสั่งซื้อทั้งหมด
	orders, err := h.store.GetOrders(c.Request.Context()) // ดึงข้อมูลทั้งหมด
	if err != nil {
		// หากเกิดข้อผิดพลาดในการดึงข้อมูลจากฐานข้อมูล
		log.Printf("Error fetching orders: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่งข้อมูลคำสั่งซื้อทั้งหมดในรูปแบบ JSON
	c.JSON(http.StatusOK, orders)
}

func (h *ProductHandlers) CreateOrder(c *gin.Context) {
	var req struct {
		CartItemIDs []int   `json:"cart_item_id"`
		TotalAmount float64 `json:"total_amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// ตรวจสอบข้อมูลเบื้องต้น
	if len(req.CartItemIDs) == 0 || req.TotalAmount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid fields"})
		return
	}

	// ตรวจสอบว่า CartItemID ทุกตัวมีค่าที่ถูกต้อง (ไม่มีค่าที่เป็น 0 หรือค่าลบ)
	var cartItems []product.CartItem
	for _, id := range req.CartItemIDs {
		if id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid cart item ID: %d", id)})
			return
		}

		// ตรวจสอบให้มั่นใจว่า CartItemID ถูกต้อง
		cartItems = append(cartItems, product.CartItem{CartItemID: id, ProductID: id})
	}

	// เรียกใช้ฟังก์ชัน CreateOrder
	orderID, err := h.store.CreateOrder(c.Request.Context(), cartItems, req.TotalAmount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create order: %v", err)})
		return
	}

	// ส่งคำตอบกลับไปยังผู้ใช้
	c.JSON(http.StatusOK, gin.H{
		"message":  "Order created successfully",
		"order_id": orderID,
	})
}

func (h *ProductHandlers) GetOrdersSort(c *gin.Context) {
	// รับค่าพารามิเตอร์สถานะจาก query string
	status := c.Param("status")

	// ถ้า status ไม่ถูกส่งมา ก็ให้ส่งข้อผิดพลาด
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status parameter is required"})
		return
	}

	// เรียกฟังก์ชันใน database layer เพื่อดึงข้อมูลคำสั่งซื้อที่จัดเรียงตามสถานะที่เลือก
	orders, err := h.store.GetOrdersSort(c.Request.Context(), status) // ดึงข้อมูลที่จัดเรียงตามสถานะ
	if err != nil {
		// หากเกิดข้อผิดพลาดในการดึงข้อมูลจากฐานข้อมูล
		log.Printf("Error fetching orders: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่งข้อมูลคำสั่งซื้อทั้งหมดในรูปแบบ JSON
	c.JSON(http.StatusOK, orders)
}

func (h *ProductHandlers) UpdateCartItemStatusHandler(c *gin.Context) {
	var input struct {
		OrderID  int `json:"order_id"`
		SellerID int `json:"seller_id"`
	}

	// ตรวจสอบว่าได้รับข้อมูล JSON ที่ถูกต้องหรือไม่
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Invalid input format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	// ดึงสถานะปัจจุบันของ cart item
	currentStatus, err := h.store.GetCurrentCartItemStatus(c.Request.Context(), input.OrderID, input.SellerID)
	if err != nil {
		log.Printf("Error retrieving current cart item status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ตรวจสอบสถานะถัดไป
	nextStatus := getNextStatus(currentStatus)
	if nextStatus == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid current status or no further status available"})
		return
	}

	// เรียกใช้ฟังก์ชันจาก database layer เพื่ออัปเดตสถานะ
	err = h.store.UpdateCartItemStatus(c.Request.Context(), input.OrderID, input.SellerID, nextStatus)
	if err != nil {
		log.Printf("Error updating cart item status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cart item status updated successfully", "new_status": nextStatus})
}

// ฟังก์ชันเพื่อระบุสถานะถัดไป
func getNextStatus(currentStatus string) string {
	statusSequence := []string{"processing", "shipping", "delivered", "received"}
	for i, status := range statusSequence {
		if status == currentStatus && i+1 < len(statusSequence) {
			return statusSequence[i+1]
		}
	}
	return ""
}

func (h *ProductHandlers) UpdateUserContactHandler(c *gin.Context) {
	// อ่านข้อมูลจาก Body
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		DisplayName string `json:"display_name"`
		Address     string `json:"address"`
		Phone       string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// ตรวจสอบว่า user_id เป็น UUID ที่ถูกต้อง
	if _, err := uuid.Parse(req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// อัปเดตข้อมูลในฐานข้อมูล
	if err := h.store.UpdateUserContact(c.Request.Context(), req.UserID, req.DisplayName, req.Address, req.Phone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update contact: %v", err)})
		return
	}

	// ส่งการตอบกลับเมื่อสำเร็จ
	c.JSON(http.StatusOK, gin.H{"message": "User contact updated successfully"})
}
