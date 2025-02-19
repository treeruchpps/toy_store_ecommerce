package product

import (
	"context"
	"database/sql"
	"fmt"

	// "log"

	"time"

	_ "github.com/lib/pq"
)

// Struct สำหรับข้อมูลสินค้า
type ProductItem struct {
	ID               int       `json:"product_id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Brand            string    `json:"brand"`
	Price            float64   `json:"price"`
	ProductStatus    string    `json:"product_status"`    // สถานะของสินค้า (In stock / No stock)
	ProductRecommend string    `json:"product_recommend"` // คำแนะนำของสินค้า (recommend / notrecommend)
	Discount         int       `json:"discount"`          // ส่วนลดสินค้า
	SellerID         int       `json:"seller_id"`
	CategoryID       int       `json:"category_id"`
	Image            string    `json:"image_url"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	Categories Category  `json:"categories"` // หมวดหมู่ของสินค้า
	Seller     Seller    `json:"seller"`     // ข้อมูลผู้ขาย
	Inventory  Inventory `json:"inventory"`  // ข้อมูลของสินค้าคงคลัง
}

// Struct สำหรับข้อมูลหมวดหมู่
type Category struct {
	ID          int    `json:"category_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Struct สำหรับข้อมูลผู้ขาย
type Seller struct {
	ID          int       `json:"seller_id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
	Description string    `json:"description"`
	Products    []Product `json:"products"` // This field holds the products of the seller
}

// Struct สำหรับข้อมูลสินค้าคงคลัง
type Inventory struct {
	Quantity  int       `json:"quantity"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Product struct สำหรับข้อมูลสินค้า
type Product struct {
	ID               int       `json:"product_id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Price            float64   `json:"price"`
	ProductStatus    string    `json:"product_status"`
	ProductRecommend string    `json:"product_recommend"`
	Discount         float64   `json:"discount"`
	Image            string    `json:"image_url"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Categories       Category  `json:"category"` // ข้อมูลหมวดหมู่ของสินค้า
}

type UpdateProduct struct {
	Price  float64 `json:"price"`
	Status string  `json:"status"`
}

type CartItem struct {
	CartItemID int           `json:"cart_item_id"`
	ProductID  int           `json:"product_id"`
	Quantity   int           `json:"quantity"`
	TotalPrice float64       `json:"total_price"`
	AddedAt    time.Time     `json:"added_at"`
	Status     string        `json:"status"`
	Product    []ProductItem `json:"product"` // เปลี่ยนเป็น array ของ ProductItem
}

type User struct {
	UserID            string     `json:"user_id"`
	GoogleID          string     `json:"google_id"`
	Email             string     `json:"email"`
	FullName          string     `json:"full_name"`
	DisplayName       string     `json:"display_name"`
	Address           *string    `json:"address"`
	Phone             *string    `json:"phone"`
	ProfilePictureURL string     `json:"profile_picture_url"`
	EmailVerified     bool       `json:"email_verified"`
	LastLoginAt       *time.Time `json:"last_login_at"`
	Status            string     `json:"status"`
	Role              string     `json:"role"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Order struct {
	OrderID     int        `json:"order_id"`
	CartItems   []CartItem `json:"cart_item_id"`
	TotalAmount float64    `json:"total_amount"`
	OrderDate   time.Time  `json:"order_date"`
}

type EcommerceDatabase interface {
	GetProduct(ctx context.Context, id string) (ProductItem, error)
	GetProductRecommend(ctx context.Context) ([]ProductItem, error)
	GetNewProducts(ctx context.Context) ([]ProductItem, error)
	SearchProducts(ctx context.Context, query string) ([]ProductItem, error)
	AllProducts(ctx context.Context) ([]ProductItem, error)
	GetSeller(ctx context.Context, id string) (Seller, error)
	GetProductByCategory(ctx context.Context, categoryID string) ([]ProductItem, error)
	AddToCart(ctx context.Context, productID, quantity int) error
	GetAllCartItems(ctx context.Context) ([]CartItem, error)
	GetUserByID(ctx context.Context, userID string) (*User, error)
	UpdateCartItemQuantity(ctx context.Context, cartItemID string, quantity int) error
	DeleteCartItem(ctx context.Context, cartItemID string) error
	CreateOrder(ctx context.Context, cartItemID []CartItem, totalAmount float64) (int, error)
	GetOrders(ctx context.Context) ([]Order, error)
	UpdateCartItemStatus(ctx context.Context, orderID int, sellerID int, status string) error
	GetOrdersSort(ctx context.Context, status string) ([]Order, error)
	GetCurrentCartItemStatus(ctx context.Context, orderID int, sellerID int) (string, error)
	UpdateUserContact(ctx context.Context, userID string, displayName, address, phone string) error
	Close() error
	Ping() error
	Reconnect(connStr string) error
}

type PostgresDatabase struct {
	db *sql.DB
}

func NewPostgresDatabase(connStr string) (*PostgresDatabase, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return &PostgresDatabase{db: db}, nil
}

// ฟังก์ชัน GetSeller สำหรับดึงข้อมูลร้านค้าตาม seller_id
func (pdb *PostgresDatabase) GetSeller(ctx context.Context, sellerID string) (Seller, error) {
	var seller Seller
	var products []Product

	// ดึงข้อมูลผู้ขาย (Seller)
	err := pdb.db.QueryRowContext(ctx, `
	SELECT seller_id, name, address, phone_number, email, description
	FROM sellers
	WHERE seller_id = $1
	`, sellerID).Scan(
		&seller.ID, &seller.Name, &seller.Address, &seller.PhoneNumber, &seller.Email, &seller.Description,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Seller{}, fmt.Errorf("seller not found")
		}
		return Seller{}, fmt.Errorf("failed to get seller: %v", err)
	}

	// ดึงข้อมูลสินค้าทั้งหมดของร้านค้าตาม seller_id
	rows, err := pdb.db.QueryContext(ctx, `
	SELECT p.product_id, p.name, p.description, p.price, p.product_status, p.product_recommend, 
		   p.discount, p.image_url, p.created_at, p.updated_at, c.category_id, c.name as category_name
	FROM products p
	LEFT JOIN categories c ON p.category_id = c.category_id
	WHERE p.seller_id = $1
	`, sellerID)
	if err != nil {
		return Seller{}, fmt.Errorf("failed to get products for seller: %v", err)
	}
	defer rows.Close()

	// ดึงข้อมูลสินค้าทั้งหมดของผู้ขาย
	for rows.Next() {
		var product Product
		var category Category
		if err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price, &product.ProductStatus, &product.ProductRecommend,
			&product.Discount, &product.Image, &product.CreatedAt, &product.UpdatedAt,
			&category.ID, &category.Name,
		); err != nil {
			return Seller{}, fmt.Errorf("failed to scan product: %v", err)
		}
		product.Categories = category
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return Seller{}, fmt.Errorf("failed to iterate over products: %v", err)
	}

	// กำหนดให้ข้อมูลสินค้าของร้านค้า
	seller.Products = products

	return seller, nil
}

func (pdb *PostgresDatabase) GetProduct(ctx context.Context, id string) (ProductItem, error) {
	var product ProductItem
	var category Category
	var seller Seller

	// ดึงข้อมูลหลักของสินค้า หมวดหมู่ และข้อมูลผู้ขาย
	err := pdb.db.QueryRowContext(ctx, `
		SELECT p.product_id, p.name, p.description, p.brand, p.price, 
		        p.seller_id, p.discount,p.image_url ,p.product_status,product_recommend,p.created_at, p.updated_at,
		       c.category_id, c.name as category_name,
		       s.seller_id, s.name as seller_name, s.address, s.phone_number, s.email, s.description as seller_description
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.category_id
		LEFT JOIN sellers s ON p.seller_id = s.seller_id
		WHERE p.product_id = $1
	`, id).Scan(
		&product.ID, &product.Name, &product.Description, &product.Brand, &product.Price,
		&product.SellerID, &product.Discount, &product.Image, &product.ProductStatus, &product.ProductRecommend, &product.CreatedAt, &product.UpdatedAt,
		&category.ID, &category.Name,
		&seller.ID, &seller.Name, &seller.Address, &seller.PhoneNumber, &seller.Email, &seller.Description)

	if err != nil {
		if err == sql.ErrNoRows {
			return ProductItem{}, fmt.Errorf("product not found")
		}
		return ProductItem{}, fmt.Errorf("failed to get product: %v", err)
	}

	product.Categories = category
	product.Seller = seller

	// ดึงข้อมูล inventory
	err = pdb.db.QueryRowContext(ctx, `
		SELECT quantity, updated_at
		FROM inventory
		WHERE product_id = $1
	`, id).Scan(&product.Inventory.Quantity, &product.Inventory.UpdatedAt)

	if err != nil && err != sql.ErrNoRows {
		return ProductItem{}, fmt.Errorf("failed to get inventory: %v", err)
	}

	return product, nil
}

func (pdb *PostgresDatabase) GetProductRecommend(ctx context.Context) ([]ProductItem, error) {
	var products []ProductItem

	rows, err := pdb.db.QueryContext(ctx, `
		SELECT p.product_id, p.name, p.description, p.brand, p.price, 
		       p.seller_id, p.discount, p.image_url, p.created_at, p.updated_at,
		       c.category_id, c.name as category_name,
		       s.seller_id, s.name as seller_name, s.address, s.phone_number, s.email, s.description as seller_description,
		       i.quantity, i.updated_at as inventory_updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.category_id
		LEFT JOIN sellers s ON p.seller_id = s.seller_id
		LEFT JOIN inventory i ON p.product_id = i.product_id
		WHERE p.product_recommend = 'recommend'
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to query recommended products: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product ProductItem
		var category Category
		var seller Seller

		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Brand, &product.Price,
			&product.SellerID, &product.Discount, &product.Image, &product.CreatedAt, &product.UpdatedAt,
			&category.ID, &category.Name,
			&seller.ID, &seller.Name, &seller.Address, &seller.PhoneNumber, &seller.Email, &seller.Description,
			&product.Inventory.Quantity, &product.Inventory.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %v", err)
		}

		product.Categories = category
		product.Seller = seller
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product rows: %v", err)
	}

	return products, nil
}

func (pdb *PostgresDatabase) GetProductByCategory(ctx context.Context, categoryID string) ([]ProductItem, error) {
	var products []ProductItem

	rows, err := pdb.db.QueryContext(ctx, `
		SELECT p.product_id, p.name, p.description, p.brand, p.price, 
		       p.seller_id, p.discount, p.image_url, p.created_at, p.updated_at,
		       c.category_id, c.name as category_name,
		       s.seller_id, s.name as seller_name, s.address, s.phone_number, s.email, s.description as seller_description,
		       i.quantity, i.updated_at as inventory_updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.category_id
		LEFT JOIN sellers s ON p.seller_id = s.seller_id
		LEFT JOIN inventory i ON p.product_id = i.product_id
		WHERE p.category_id = $1
	`, categoryID)

	if err != nil {
		return nil, fmt.Errorf("failed to query products by category: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product ProductItem
		var category Category
		var seller Seller

		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Brand, &product.Price,
			&product.SellerID, &product.Discount, &product.Image, &product.CreatedAt, &product.UpdatedAt,
			&category.ID, &category.Name,
			&seller.ID, &seller.Name, &seller.Address, &seller.PhoneNumber, &seller.Email, &seller.Description,
			&product.Inventory.Quantity, &product.Inventory.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %v", err)
		}

		product.Categories = category
		product.Seller = seller
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product rows: %v", err)
	}

	return products, nil
}

func (pdb *PostgresDatabase) GetNewProducts(ctx context.Context) ([]ProductItem, error) {
	var products []ProductItem

	rows, err := pdb.db.QueryContext(ctx, `
        SELECT DISTINCT ON (p.seller_id) p.product_id, p.name, p.description, p.brand, p.price, 
               p.seller_id, p.discount, p.image_url, p.created_at, p.updated_at,
               c.category_id, c.name as category_name,
               s.seller_id, s.name as seller_name, s.address, s.phone_number, s.email, s.description as seller_description,
               i.quantity, i.updated_at as inventory_updated_at
        FROM products p
        LEFT JOIN categories c ON p.category_id = c.category_id
        LEFT JOIN sellers s ON p.seller_id = s.seller_id
        LEFT JOIN inventory i ON p.product_id = i.product_id
        ORDER BY p.seller_id, p.created_at DESC
    `)

	if err != nil {
		return nil, fmt.Errorf("failed to query new products by seller: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product ProductItem
		var category Category
		var seller Seller

		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Brand, &product.Price,
			&product.SellerID, &product.Discount, &product.Image, &product.CreatedAt, &product.UpdatedAt,
			&category.ID, &category.Name,
			&seller.ID, &seller.Name, &seller.Address, &seller.PhoneNumber, &seller.Email, &seller.Description,
			&product.Inventory.Quantity, &product.Inventory.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %v", err)
		}

		product.Categories = category
		product.Seller = seller
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product rows: %v", err)
	}

	return products, nil
}

func (pdb *PostgresDatabase) AllProducts(ctx context.Context) ([]ProductItem, error) {
	var products []ProductItem

	// Query to fetch all products from the database
	rows, err := pdb.db.QueryContext(ctx, `
		SELECT p.product_id, p.name, p.description, p.brand, p.price, 
		       p.seller_id, p.discount, p.image_url, p.created_at, p.updated_at,
		       c.category_id, c.name as category_name,
		       s.seller_id, s.name as seller_name, s.address, s.phone_number, s.email, s.description as seller_description,
		       i.quantity, i.updated_at as inventory_updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.category_id
		LEFT JOIN sellers s ON p.seller_id = s.seller_id
		LEFT JOIN inventory i ON p.product_id = i.product_id ORDER BY product_id ASC;
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to query all products: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product ProductItem
		var category Category
		var seller Seller

		// Scanning each row into the product struct and its related data
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Brand, &product.Price,
			&product.SellerID, &product.Discount, &product.Image, &product.CreatedAt, &product.UpdatedAt,
			&category.ID, &category.Name,
			&seller.ID, &seller.Name, &seller.Address, &seller.PhoneNumber, &seller.Email, &seller.Description,
			&product.Inventory.Quantity, &product.Inventory.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %v", err)
		}

		// Add product with its category and seller info to the products slice
		product.Categories = category
		product.Seller = seller
		products = append(products, product)
	}

	// Check for any errors that occurred during the iteration of rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product rows: %v", err)
	}

	// Return the list of products
	return products, nil
}

func (pdb *PostgresDatabase) SearchProducts(ctx context.Context, query string) ([]ProductItem, error) {
	var products []ProductItem

	// Corrected SQL query with placeholders for parameters
	rows, err := pdb.db.QueryContext(ctx, `
		SELECT p.product_id, p.name, p.description, p.brand, p.price, 
		       p.seller_id, p.discount, p.image_url, p.created_at, p.updated_at,
		       c.category_id, c.name as category_name,
		       s.seller_id, s.name as seller_name, s.address, s.phone_number, s.email, s.description as seller_description,
		       i.quantity, i.updated_at as inventory_updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.category_id
		LEFT JOIN sellers s ON p.seller_id = s.seller_id
		LEFT JOIN inventory i ON p.product_id = i.product_id
		WHERE p.name ILIKE $1
	`, "%"+query+"%") // Passing query parameter safely

	if err != nil {
		return nil, fmt.Errorf("failed to search products: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product ProductItem
		var category Category
		var seller Seller
		// You should also scan the additional fields into the appropriate structs if needed
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Brand, &product.Price,
			&product.SellerID, &product.Discount, &product.Image, &product.CreatedAt, &product.UpdatedAt,
			&category.ID, &category.Name,
			&seller.ID, &seller.Name, &seller.Address, &seller.PhoneNumber, &seller.Email, &seller.Description,
			&product.Inventory.Quantity, &product.Inventory.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %v", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product rows: %v", err)
	}

	return products, nil
}

func (pdb *PostgresDatabase) AddToCart(ctx context.Context, productID, quantity int) error {
	// ตรวจสอบว่ามีสินค้ารายการนี้อยู่ในฐานข้อมูลและดึงราคาของสินค้า
	var price float64
	err := pdb.db.QueryRowContext(ctx, `SELECT price FROM products WHERE product_id = $1`, productID).Scan(&price)
	if err != nil {
		return fmt.Errorf("failed to get product price: %v", err)
	}

	// คำนวณ total_price
	totalPrice := float64(quantity) * price

	// ตรวจสอบว่ามีสินค้านี้อยู่แล้วในตะกร้า ถ้ามีแล้วให้เพิ่มจำนวนสินค้าและอัปเดต total_price
	var existsInCart bool
	err = pdb.db.QueryRowContext(ctx, `
        SELECT EXISTS(SELECT 1 FROM cart_items WHERE product_id = $1)
    `, productID).Scan(&existsInCart)
	if err != nil {
		return fmt.Errorf("failed to check if product exists in cart: %v", err)
	}

	if existsInCart {
		// ตรวจสอบว่า add_to_cart เป็น true หรือไม่
		var addToCart bool
		err = pdb.db.QueryRowContext(ctx, `
			SELECT added_to_cart
			FROM cart_items
			WHERE product_id = $1
		`, productID).Scan(&addToCart)
		if err != nil {
			return fmt.Errorf("failed to check add_to_cart status: %v", err)
		}

		if addToCart {
			// สร้างตะกร้าใหม่หาก add_to_cart เป็น true
			_, err = pdb.db.ExecContext(ctx, `
				INSERT INTO cart_items (product_id, quantity, total_price, added_at)
				VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
			`, productID, quantity, totalPrice)
			if err != nil {
				return fmt.Errorf("failed to create a new cart: %v", err)
			}
		} else {
			// อัปเดตจำนวนสินค้าที่มีอยู่ในตะกร้าเดิม
			_, err = pdb.db.ExecContext(ctx, `
				UPDATE cart_items
				SET quantity = quantity + $1, total_price = total_price + $2
				WHERE product_id = $3
			`, quantity, totalPrice, productID)
			if err != nil {
				return fmt.Errorf("failed to update product quantity in cart: %v", err)
			}
		}
	} else {
		// เพิ่มสินค้ารายการใหม่ในตะกร้า
		_, err = pdb.db.ExecContext(ctx, `
			INSERT INTO cart_items (product_id, quantity, total_price, added_at)
			VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		`, productID, quantity, totalPrice)
		if err != nil {
			return fmt.Errorf("failed to add product to cart: %v", err)
		}
	}

	return nil
}

func (pdb *PostgresDatabase) GetAllCartItems(ctx context.Context) ([]CartItem, error) {
	query := `SELECT ci.cart_item_id, ci.product_id, ci.quantity, ci.added_at, ci.status,
                      p.product_id, p.name, p.description, p.price, 
                      (p.price * ci.quantity) AS total_price, 
                      p.product_status, p.product_recommend, p.discount, p.image_url, 
                      p.created_at, p.updated_at,
                      c.category_id, c.name AS category_name, c.description AS category_description,
                      s.seller_id, s.name AS seller_name, s.address, s.phone_number, s.email, s.description AS seller_description,
                      i.quantity AS inventory_quantity, i.updated_at AS inventory_updated_at
               FROM cart_items ci
               JOIN products p ON ci.product_id = p.product_id
               LEFT JOIN categories c ON p.category_id = c.category_id
               LEFT JOIN sellers s ON p.seller_id = s.seller_id
               LEFT JOIN inventory i ON p.product_id = i.product_id
               WHERE ci.added_to_cart = FALSE
               ORDER BY ci.cart_item_id
`
	rows, err := pdb.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var cartItems []CartItem

	// Loop through the rows
	for rows.Next() {
		var cartItem CartItem
		var productItem ProductItem
		var category Category
		var seller Seller
		var inventory Inventory

		// Scan data from the database
		err := rows.Scan(
			&cartItem.CartItemID, &cartItem.ProductID, &cartItem.Quantity, &cartItem.AddedAt, &cartItem.Status,
			&productItem.ID, &productItem.Name, &productItem.Description, &productItem.Price, &cartItem.TotalPrice,
			&productItem.ProductStatus, &productItem.ProductRecommend, &productItem.Discount, &productItem.Image,
			&productItem.CreatedAt, &productItem.UpdatedAt,
			&category.ID, &category.Name, &category.Description,
			&seller.ID, &seller.Name, &seller.Address, &seller.PhoneNumber, &seller.Email, &seller.Description,
			&inventory.Quantity, &inventory.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Set productItem details from the database
		productItem.Categories = category
		productItem.Seller = seller
		productItem.Inventory = inventory

		// Append the productItem to the CartItem's Product slice
		cartItem.Product = append(cartItem.Product, productItem)

		// Add the CartItem to the cartItems slice
		cartItems = append(cartItems, cartItem)
	}

	// Check if there were no rows
	if len(cartItems) == 0 {
		return nil, fmt.Errorf("no cart items found")
	}

	return cartItems, nil
}

func (pdb *PostgresDatabase) UpdateCartItemQuantity(ctx context.Context, cartItemID string, quantity int) error {
	// ตรวจสอบว่ามีรายการสินค้านี้ในตะกร้าหรือไม่
	var existsInCart bool
	err := pdb.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM cart_items WHERE cart_item_id = $1)`, cartItemID).Scan(&existsInCart)
	if err != nil {
		return fmt.Errorf("failed to check if cart item exists: %v", err)
	}

	if !existsInCart {
		return fmt.Errorf("cart item with ID '%s' not found", cartItemID)
	}

	// ตรวจสอบสินค้าคงเหลือในตาราง inventory
	var stockQuantity int
	var price float64
	err = pdb.db.QueryRowContext(ctx, `SELECT quantity, price FROM inventory i JOIN products p ON p.product_id = i.product_id WHERE i.product_id = (SELECT product_id FROM cart_items WHERE cart_item_id = $1)`, cartItemID).Scan(&stockQuantity, &price)
	if err != nil {
		return fmt.Errorf("failed to check stock and price for product in inventory: %v", err)
	}

	// หากจำนวนที่ต้องการอัปเดตมากกว่าสินค้าคงเหลือให้ใช้จำนวนสินค้าคงเหลือมากที่สุด
	if quantity > stockQuantity {
		quantity = stockQuantity
	}

	// คำนวณราคาใหม่
	totalPrice := price * float64(quantity)

	// อัปเดตจำนวนสินค้าและราคาสินค้าในตะกร้า
	_, err = pdb.db.ExecContext(ctx, `UPDATE cart_items SET quantity = $1, total_price = $2 WHERE cart_item_id = $3`, quantity, totalPrice, cartItemID)
	if err != nil {
		return fmt.Errorf("failed to update cart item quantity and price: %v", err)
	}

	return nil
}

func (pdb *PostgresDatabase) DeleteCartItem(ctx context.Context, cartItemID string) error {
	// ตรวจสอบการมีอยู่ของรายการในตะกร้าก่อนลบ
	var existsInCart bool
	err := pdb.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM cart_items WHERE cart_item_id = $1 )`, cartItemID).Scan(&existsInCart)
	if err != nil {
		return fmt.Errorf("failed to check if cart item exists: %v", err)
	}

	if !existsInCart {
		// ข้อความนี้จะบอกว่าไม่พบรายการสินค้าตาม ID ที่ให้มา
		return fmt.Errorf("cart item with ID '%s' not found for deletion", cartItemID)
	}

	// ลบสินค้าจากตะกร้าโดยใช้ cart_item_id
	_, err = pdb.db.ExecContext(ctx, `DELETE FROM cart_items WHERE cart_item_id = $1`, cartItemID)
	if err != nil {
		return fmt.Errorf("failed to delete cart item: %v", err)
	}

	return nil
}

func (pdb *PostgresDatabase) Close() error {
	return pdb.db.Close()
}

func (pdb *PostgresDatabase) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return pdb.db.PingContext(ctx)
}

func (pdb *PostgresDatabase) Reconnect(connStr string) error {
	if pdb.db != nil {
		pdb.db.Close()
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// ตั้งค่า connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	pdb.db = db
	return nil
}

func (pdb *PostgresDatabase) GetUserByID(ctx context.Context, userID string) (*User, error) {
	var user User

	// คำสั่ง SQL ที่ดึงข้อมูลผู้ใช้จากฐานข้อมูล
	err := pdb.db.QueryRowContext(ctx, `
        SELECT user_id, google_id, email, display_name, address, phone, profile_picture_url,
               email_verified, last_login_at, status, role, created_at, updated_at
        FROM users
        WHERE user_id = $1
    `, userID).Scan(
		&user.UserID, &user.GoogleID, &user.Email, &user.DisplayName, &user.Address, &user.Phone, &user.ProfilePictureURL,
		&user.EmailVerified, &user.LastLoginAt, &user.Status, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)

	// ตรวจสอบข้อผิดพลาดในการดึงข้อมูล
	if err != nil {
		// ถ้าไม่พบผู้ใช้
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with ID %s not found", userID)
		}
		// ถ้ามีข้อผิดพลาดจากฐานข้อมูล
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	// คืนค่า pointer ไปยัง User
	return &user, nil
}

func (pdb *PostgresDatabase) CreateOrder(ctx context.Context, cartItems []CartItem, totalAmount float64) (int, error) {
	var orderID int

	// เริ่มต้น transaction
	tx, err := pdb.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %v", err)
	}

	// คำสั่ง SQL สำหรับการสร้างคำสั่งซื้อ
	stmt := `INSERT INTO orders (total_amount) 
          VALUES ($1) RETURNING order_id`
	err = tx.QueryRowContext(ctx, stmt, totalAmount).Scan(&orderID)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to create order: %v", err)
	}

	// เพิ่ม CartItems ในคำสั่งซื้อ
	for _, item := range cartItems {
		// ตรวจสอบว่า ProductID ของแต่ละ CartItem ไม่เป็น 0 หรือค่าผิดปกติ
		if item.ProductID <= 0 {
			tx.Rollback()
			return 0, fmt.Errorf("invalid product ID %d for cart item", item.ProductID)
		}

		// ดึงข้อมูล product_id จาก cart_items
		var productID int
		err := tx.QueryRowContext(ctx, `SELECT product_id FROM cart_items WHERE cart_item_id = $1`, item.CartItemID).Scan(&productID)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to fetch product_id for cart_item_id %d: %v", item.CartItemID, err)
		}

		// ดึงข้อมูล seller_id ของสินค้า
		var productSellerID int
		err = tx.QueryRowContext(ctx, `SELECT seller_id FROM products WHERE product_id = $1`, productID).Scan(&productSellerID)
		if err == sql.ErrNoRows {
			tx.Rollback()
			return 0, fmt.Errorf("no seller found for product_id %d", productID)
		} else if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to fetch seller ID for product %d: %v", productID, err)
		}

		// คำสั่ง SQL สำหรับการเพิ่ม CartItem ลงใน order_items
		_, err = tx.ExecContext(ctx, `INSERT INTO order_items (order_id, cart_item_id, seller_id) 
                                    VALUES ($1, $2, $3)`, orderID, item.CartItemID, productSellerID)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to add cart item to order: %v", err)
		}
	}

	// คำสั่ง SQL สำหรับการอัปเดตสถานะ added_to_cart
	_, err = tx.ExecContext(ctx, `UPDATE cart_items SET added_to_cart = TRUE WHERE added_to_cart = FALSE`)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to update cart items: %v", err)
	}

	// ยืนยันการทำธุรกรรม
	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return orderID, nil
}

func (pdb *PostgresDatabase) GetOrders(ctx context.Context) ([]Order, error) {
	query := `
        SELECT 
            o.order_id, o.total_amount, o.order_date, 
            COALESCE(ci.cart_item_id, 0) AS cart_item_id, ci.product_id, ci.quantity, ci.total_price, ci.added_at, ci.status,
            p.product_id, p.name AS product_name, p.description AS product_description, 
            p.price, p.product_status, p.product_recommend, p.discount, p.image_url, 
            c.category_id, COALESCE(c.name, 'No Category') AS category_name,
            s.seller_id, COALESCE(s.name, 'Unknown Seller') AS seller_name,
            COALESCE(s.address, 'No Address') AS seller_address,
            COALESCE(s.phone_number, 'No Phone') AS seller_phone,
            COALESCE(i.quantity, 0) AS inventory_quantity
        FROM orders o
        LEFT JOIN order_items oi ON o.order_id = oi.order_id
        LEFT JOIN cart_items ci ON oi.cart_item_id = ci.cart_item_id
        LEFT JOIN products p ON ci.product_id = p.product_id
        LEFT JOIN categories c ON p.category_id = c.category_id
        LEFT JOIN sellers s ON p.seller_id = s.seller_id
        LEFT JOIN inventory i ON p.product_id = i.product_id
        ORDER BY o.order_id DESC, ci.cart_item_id DESC;
    `

	rows, err := pdb.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}
	defer rows.Close()

	ordersMap := make(map[int]*Order)
	for rows.Next() {
		var orderID int
		var order Order
		var cartItem CartItem
		var productItem ProductItem
		var category Category
		var seller Seller

		err := rows.Scan(
			&orderID, &order.TotalAmount, &order.OrderDate,
			&cartItem.CartItemID, &cartItem.ProductID, &cartItem.Quantity, &cartItem.TotalPrice, &cartItem.AddedAt, &cartItem.Status,
			&productItem.ID, &productItem.Name, &productItem.Description, &productItem.Price,
			&productItem.ProductStatus, &productItem.ProductRecommend, &productItem.Discount, &productItem.Image,
			&category.ID, &category.Name,
			&seller.ID, &seller.Name, &seller.Address, &seller.PhoneNumber,
			&productItem.Inventory.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// ตรวจสอบการใช้ Null หรือค่าที่อาจเป็น NULL
		if cartItem.CartItemID == 0 {
			cartItem.CartItemID = 0 // หรือสามารถปรับให้เป็นค่า default อื่นๆ ตามต้องการ
		}

		if productItem.ID == 0 {
			productItem.ID = 0 // หรือค่า default
		}

		// สร้าง ProductItem และเพิ่มข้อมูล
		productItem.Categories = category
		productItem.Seller = seller

		// เพิ่ม ProductItem ใน CartItem
		cartItem.Product = []ProductItem{productItem}

		// ตรวจสอบว่า Order มีอยู่ใน Map หรือยัง
		if existingOrder, exists := ordersMap[orderID]; exists {
			// ถ้ามี order นี้อยู่แล้ว ให้เพิ่ม cartItem ใหม่ใน CartItems
			existingOrder.CartItems = append(existingOrder.CartItems, cartItem)
		} else {
			// ถ้ายังไม่มี ให้สร้าง Order ใหม่และเพิ่ม CartItems ใหม่
			order.OrderID = orderID
			order.CartItems = []CartItem{cartItem}
			ordersMap[orderID] = &order
		}
	}

	// ตรวจสอบข้อผิดพลาดขณะ iterate rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Convert map to slice
	var orders []Order
	for _, order := range ordersMap {
		orders = append(orders, *order)
	}

	return orders, nil
}

func (pdb *PostgresDatabase) UpdateCartItemStatus(ctx context.Context, orderID int, sellerID int, status string) error {
	// อัปเดตสถานะในตาราง cart_items โดยอ้างอิงจาก order_id และ seller_id
	_, err := pdb.db.ExecContext(ctx, `
        UPDATE cart_items 
        SET status = $1
        WHERE cart_item_id IN (
            SELECT cart_item_id 
            FROM order_items 
            WHERE order_id = $2 AND seller_id = $3
        )`, status, orderID, sellerID)
	if err != nil {
		return fmt.Errorf("failed to update cart item status: %v", err)
	}
	return nil
}

func (pdb *PostgresDatabase) UpdateUserContact(ctx context.Context, userID string, displayName, address, phone string) error {
	query := `
        UPDATE users
        SET 
            display_name = COALESCE(NULLIF($1, ''), display_name), 
            address = COALESCE(NULLIF($2, ''), address), 
            phone = COALESCE(NULLIF($3, ''), phone), 
            updated_at = CURRENT_TIMESTAMP
        WHERE user_id = $4
    `
	_, err := pdb.db.ExecContext(ctx, query, displayName, address, phone, userID)
	if err != nil {
		return fmt.Errorf("failed to update user contact: %v", err)
	}
	return nil
}

func (pdb *PostgresDatabase) GetOrdersSort(ctx context.Context, status string) ([]Order, error) {
	query := `
        SELECT 
            o.order_id, o.total_amount, o.order_date, 
            COALESCE(ci.cart_item_id, 0) AS cart_item_id, ci.product_id, ci.quantity, ci.total_price, ci.added_at, ci.status,
            p.product_id, p.name AS product_name, p.description AS product_description, 
            p.price, p.product_status, p.product_recommend, p.discount, p.image_url, 
            c.category_id, COALESCE(c.name, 'No Category') AS category_name,
            s.seller_id, COALESCE(s.name, 'Unknown Seller') AS seller_name,
            COALESCE(s.address, 'No Address') AS seller_address,
            COALESCE(s.phone_number, 'No Phone') AS seller_phone,
            COALESCE(i.quantity, 0) AS inventory_quantity
        FROM orders o
        LEFT JOIN order_items oi ON o.order_id = oi.order_id
        LEFT JOIN cart_items ci ON oi.cart_item_id = ci.cart_item_id
        LEFT JOIN products p ON ci.product_id = p.product_id
        LEFT JOIN categories c ON p.category_id = c.category_id
        LEFT JOIN sellers s ON p.seller_id = s.seller_id
        LEFT JOIN inventory i ON p.product_id = i.product_id
    `

	// กรองคำสั่งซื้อที่มีสถานะที่ตรงกับที่ผู้ใช้ระบุ
	if status != "" {
		query += ` WHERE ci.status = $1`
	}

	// เพิ่มการจัดเรียงผลลัพธ์
	query += ` ORDER BY o.order_id DESC, ci.cart_item_id DESC;`

	// เรียกใช้คำสั่ง SQL พร้อมค่าพารามิเตอร์
	rows, err := pdb.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}
	defer rows.Close()

	ordersMap := make(map[int]*Order)
	for rows.Next() {
		var orderID int
		var order Order
		var cartItem CartItem
		var productItem ProductItem
		var category Category
		var seller Seller

		err := rows.Scan(
			&orderID, &order.TotalAmount, &order.OrderDate,
			&cartItem.CartItemID, &cartItem.ProductID, &cartItem.Quantity, &cartItem.TotalPrice, &cartItem.AddedAt, &cartItem.Status,
			&productItem.ID, &productItem.Name, &productItem.Description, &productItem.Price,
			&productItem.ProductStatus, &productItem.ProductRecommend, &productItem.Discount, &productItem.Image,
			&category.ID, &category.Name,
			&seller.ID, &seller.Name, &seller.Address, &seller.PhoneNumber,
			&productItem.Inventory.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// สร้าง ProductItem และเพิ่มข้อมูล
		productItem.Categories = category
		productItem.Seller = seller

		// เพิ่ม ProductItem ใน CartItem
		cartItem.Product = []ProductItem{productItem}

		// ตรวจสอบว่า Order มีอยู่ใน Map หรือยัง
		if existingOrder, exists := ordersMap[orderID]; exists {
			// ถ้ามี order นี้อยู่แล้ว ให้เพิ่ม cartItem ใหม่ใน CartItems
			existingOrder.CartItems = append(existingOrder.CartItems, cartItem)
		} else {
			// ถ้ายังไม่มี ให้สร้าง Order ใหม่และเพิ่ม CartItems ใหม่
			order.OrderID = orderID
			order.CartItems = []CartItem{cartItem}
			ordersMap[orderID] = &order
		}
	}

	// ตรวจสอบข้อผิดพลาดขณะ iterate rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	var orders []Order
	for _, order := range ordersMap {
		orders = append(orders, *order)
	}

	return orders, nil
}

func (pdb *PostgresDatabase) GetCurrentCartItemStatus(ctx context.Context, orderID int, sellerID int) (string, error) {
	var currentStatus string
	err := pdb.db.QueryRowContext(ctx, `
        SELECT ci.status
        FROM cart_items ci
        JOIN order_items oi ON ci.cart_item_id = oi.cart_item_id
        WHERE oi.order_id = $1 AND oi.seller_id = $2
        LIMIT 1`, orderID, sellerID).Scan(&currentStatus)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no cart items found for the given order and seller")
		}
		return "", fmt.Errorf("failed to retrieve cart item status: %v", err)
	}

	return currentStatus, nil
}

type Store struct {
	db EcommerceDatabase
}

func NewStore(db EcommerceDatabase) *Store {
	return &Store{db: db}
}

func (s *Store) GetProduct(ctx context.Context, id string) (ProductItem, error) {
	return s.db.GetProduct(ctx, id)
}

func (s *Store) AllProducts(ctx context.Context) ([]ProductItem, error) {
	return s.db.AllProducts(ctx)
}

func (s *Store) GetProductRecommend(ctx context.Context) ([]ProductItem, error) {
	return s.db.GetProductRecommend(ctx)
}

func (s *Store) GetNewProducts(ctx context.Context) ([]ProductItem, error) {
	return s.db.GetNewProducts(ctx)
}

func (s *Store) SearchProducts(ctx context.Context, query string) ([]ProductItem, error) {
	return s.db.SearchProducts(ctx, query)
}

func (s *Store) GetSeller(ctx context.Context, id string) (Seller, error) {
	return s.db.GetSeller(ctx, id)
}

func (s *Store) GetProductByCategory(ctx context.Context, categoryID string) ([]ProductItem, error) {
	return s.db.GetProductByCategory(ctx, categoryID)
}

func (s *Store) AddToCart(ctx context.Context, productID, quantity int) error {
	return s.db.AddToCart(ctx, productID, quantity)
}

func (s *Store) GetAllCartItems(ctx context.Context) ([]CartItem, error) {
	return s.db.GetAllCartItems(ctx)
}

func (s *Store) UpdateCartItemQuantity(ctx context.Context, cartItemID string, quantity int) error {
	return s.db.UpdateCartItemQuantity(ctx, cartItemID, quantity)
}

func (s *Store) DeleteCartItem(ctx context.Context, cartItemID string) error {
	return s.db.DeleteCartItem(ctx, cartItemID)
}

func (s *Store) CreateOrder(ctx context.Context, cartItemID []CartItem, totalAmount float64) (int, error) {
	return s.db.CreateOrder(ctx, cartItemID, totalAmount)
}

func (s *Store) GetOrders(ctx context.Context) ([]Order, error) {
	return s.db.GetOrders(ctx)
}

func (s *Store) UpdateCartItemStatus(ctx context.Context, orderID int, sellerID int, status string) error {
	return s.db.UpdateCartItemStatus(ctx, orderID, sellerID, status)
}

func (s *Store) GetOrdersSort(ctx context.Context, status string) ([]Order, error) {
	return s.db.GetOrdersSort(ctx, status)
}

func (s *Store) GetCurrentCartItemStatus(ctx context.Context, orderID int, sellerID int) (string, error) {
	return s.db.GetCurrentCartItemStatus(ctx, orderID, sellerID)
}

func (s *Store) UpdateUserContact(ctx context.Context, userID string, displayName, address, phone string) error {
	return s.db.UpdateUserContact(ctx, userID, displayName, address, phone)
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Ping() error {
	if s.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return s.db.Ping()
}

func (s *Store) Reconnect(connStr string) error {
	return s.db.Reconnect(connStr)
}

// GetUserByID: ฟังก์ชันใน Store เพื่อเรียกใช้ PostgresDatabase
func (s *Store) GetUserByID(ctx context.Context, userID string) (*User, error) {
	return s.db.GetUserByID(ctx, userID)
}
