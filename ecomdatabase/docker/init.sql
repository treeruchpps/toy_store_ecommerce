-- สร้าง Extension สำหรับ UUID (เฉพาะ PostgreSQL)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- สร้าง ENUM สำหรับ status และ product_type
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'product_recommend') THEN
        CREATE TYPE product_recommend AS ENUM ('recommend', 'notrecommend');
    END IF;
END$$;

-- สร้างตาราง categories
CREATE TABLE IF NOT EXISTS categories (
    category_id SERIAL PRIMARY KEY,
    name VARCHAR(60) NOT NULL UNIQUE,
    description TEXT
);

-- สร้างตาราง sellers
CREATE TABLE IF NOT EXISTS sellers (
    seller_id SERIAL PRIMARY KEY,  -- ใช้ SERIAL เพื่อให้มีการสร้าง ID อัตโนมัติ
    name VARCHAR(255) NOT NULL UNIQUE,
    address VARCHAR(255),
    phone_number CHAR(10),
    email VARCHAR(50),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- สร้างตาราง products
CREATE TABLE IF NOT EXISTS products (
    product_id SERIAL PRIMARY KEY,  -- ใช้ SERIAL เพื่อให้มีการสร้าง ID อัตโนมัติ
    name VARCHAR(255) NOT NULL,
    description TEXT,
    product_stock INTEGER DEFAULT 0 CHECK (product_stock >= 0),
    brand VARCHAR(255),
    price NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
    product_status VARCHAR(20) GENERATED ALWAYS AS (
        CASE 
            WHEN product_stock > 0 THEN 'In stock'
            ELSE 'No stock'
        END
    ) STORED,
    product_recommend product_recommend NOT NULL,
    discount INTEGER DEFAULT 0,
    image_url VARCHAR(255),
    seller_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (seller_id) REFERENCES sellers(seller_id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE CASCADE
);

-- สร้างตาราง cart_items ใหม่
CREATE TABLE IF NOT EXISTS cart_items (
    cart_item_id SERIAL PRIMARY KEY, -- รหัสไอเท็มในตะกร้าเป็น UUID
    product_id INT NOT NULL,                                 -- รหัสสินค้า
    quantity INT NOT NULL CHECK (quantity > 0), 
    total_price NUMERIC(10, 2) NOT NULL,              -- จำนวนสินค้าที่เลือก
    added_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,          -- วันที่เพิ่มสินค้าลงตะกร้า
    added_to_cart BOOLEAN DEFAULT FALSE,
    status VARCHAR(50) DEFAULT 'processing',              -- สถานะของคำสั่งซื้อ 
    FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE CASCADE -- เชื่อมโยงกับตาราง products
);

CREATE TABLE IF NOT EXISTS orders (
    order_id SERIAL PRIMARY KEY,                          -- รหัสคำสั่งซื้อ
    total_amount NUMERIC(10, 2) NOT NULL,                 -- ยอดรวมของคำสั่งซื้อ
    order_date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP    -- วันที่ทำการสั่งซื้อ
);

CREATE TABLE IF NOT EXISTS order_items (
    order_item_id SERIAL PRIMARY KEY,                     -- รหัสไอเท็มในคำสั่งซื้อ
    order_id INT NOT NULL,                                -- รหัสคำสั่งซื้อ
    cart_item_id INT NOT NULL,                            -- รหัสไอเท็มในตะกร้า
    seller_id INT NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE CASCADE,
    FOREIGN KEY (cart_item_id) REFERENCES cart_items(cart_item_id) ON DELETE CASCADE,
    FOREIGN KEY (seller_id) REFERENCES sellers(seller_id) ON DELETE CASCADE
);

-- สร้างฟังก์ชันสำหรับอัปเดตฟิลด์ updated_at อัตโนมัติ
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC';
   RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

-- สร้าง Trigger สำหรับตารางที่มีฟิลด์ updated_at
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_sellers_updated_at BEFORE UPDATE ON sellers
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- แทรกข้อมูลตัวอย่างลงใน categories
INSERT INTO categories (name, description) VALUES
('แรกเกิด - 6 เดือน', 'ของเล่นสำหรับเด็กแรกเกิดจนถึง 6 เดือน'),
('7 - 11 เดือน', 'ของเล่นสำหรับเด็กช่วงอายุ 7 ถึง 11 เดือน'),
('1 - 2 ขวบ', 'ของเล่นสำหรับเด็กช่วงอายุ 1 ถึง 2 ขวบ'),
('3 ขวบขึ้นไป', 'ของเล่นสำหรับเด็กอายุ 3 ปีขึ้นไป');

-- แทรกข้อมูลตัวอย่างลงใน sellers
INSERT INTO sellers (name, address, phone_number, email, description)
VALUES
('Happy Hippo', '303 Ladkrabang Bangkok', '0112233445', 'happyhippo@gmail.com', 'ร้านขายของเล่นเสริมพัฒนาการสำหรับเด็ก พร้อมการบริการที่อบอุ่น'),
('Happy Land', '123 Bangkapi Bangkok', '0987654321', 'happyland@gmail.com', 'ร้านของเล่นเด็กที่มีสินค้าให้เลือกหลากหลาย เหมาะสำหรับเด็กทุกวัย'),
('Teddy Bear', '232 Dusit Bangkok', '0987897898', 'teddybear@gmail.com', 'ร้านจำหน่ายตุ๊กตาหมีและของเล่นน่ารักๆ สำหรับเด็กและผู้ใหญ่'),
('Fun Zone', '456 Bangna Bangkok', '0766786789', 'funzone@gmail.com', 'ร้านขายของเล่นและอุปกรณ์สำหรับเล่นสนุกทั้งในบ้านและนอกบ้าน'),
('Dino Shop', '325 Phayathai Bangkok', '0876785678', 'dinoshop@gmail.com', 'ร้านของเล่นและของสะสมเกี่ยวกับไดโนเสาร์สำหรับแฟนพันธุ์แท้และเด็ก');

INSERT INTO products (
    product_id, name, description, product_stock, price, product_recommend,
    discount, seller_id, category_id, brand, image_url
) VALUES
    (1, 'บอลตาข่ายกรุ๊งกริ้ง', 'สีสันสดใส เสียงดังฟังชัด ตาข่ายนิ่ม', 39, 109, 'recommend', 0, 1, 1, 'BrandA', 'https://aws.cmzimg.com/upload/10267/product-images/BB384923/5035b69a.jpg'),
    (2, 'ตุ๊กตากล่อมนอน', 'มีเสียงดนตรีกล่องนอน และเสียงธรรมชาติ ใช้ถ่าน AA 2 ก้อนที่กล่องดนตรี', 25, 919, 'notrecommend', 46, 1, 1, 'BrandB', 'https://aws.cmzimg.com/upload/10267/product-images/BB381982/7658f8e1.jpg'),
    (3, 'ออร์แกนเด็กรูปอมยิ้ม', 'มีเพลงกล่อมนอนสามารถเปลี่ยนเพลงได้เองไปเรื่อยๆ ใช้ถ่าน AAA 2 ก้อน', 14, 229, 'notrecommend', 0, 1, 1, 'BrandC', 'https://aws.cmzimg.com/upload/10267/product-images/BB370886/6096d566.jpg'),
    (4, 'บอลบีบฝึกกล้ามเนื้อมือ', 'บอลบีบฝึกกล้ามเนื้อมือ 1 ชุดมี 10 ลูกหลากหลายสี', 7, 519, 'notrecommend', 0, 1, 1, 'BrandD', 'https://aws.cmzimg.com/upload/10267/product-images/BB003466/9e9da7f8.jpg'),
    (5, 'โมบายผ้ารูปดาว', 'ผ้านุ่มนิ่ม สีสดใส ไม่ต้องใช้ถ่าน', 14, 129, 'notrecommend', 0, 1, 1, 'BrandE', 'https://aws.cmzimg.com/upload/10267/product-images/BB161463/544d34f7.jpg'),
    (6, 'รถเด็กหัดเดินรูปแมว', 'รถเด็กหัดเดินรูปแมว สีชมพู มีไฟ-เสียงเพลง เล่นMP3ได้ ปรับหนืดได้', 5, 1390, 'recommend', 0, 2, 2, 'BrandF', 'https://aws.cmzimg.com/upload/10267/product-images/FW-1615/f7e04910.jpg'),
    (7, 'School Bus', 'School Bus สอนภาษา', 19, 1090, 'notrecommend', 0, 2, 2, 'BrandG', 'https://aws.cmzimg.com/upload/10267/product-images/BB856403B/32b40959.jpg'),
    (8, 'ออร์แกนเสียงสัตว์', 'ใช้ถ่าน AA จำนวน 3 ก้อน', 10, 199, 'notrecommend', 40, 2, 2, 'BrandH', 'https://aws.cmzimg.com/upload/10267/product-images/BB808166E-WL/e04cc1db.jpg'),
    (9, 'ออร์แกนรูปช้าง', 'ใช้ถ่าน AA 3 ก้อน มีเสียงสัตว์ เปลี่ยนโหมดได้', 11, 199, 'notrecommend', 0, 2, 2, 'BrandI', 'https://aws.cmzimg.com/upload/10267/product-images/BB993136W/ea28b0c0.jpg'),
    (10, 'บล็อกหยอดรูปทรงต่างๆ', 'ขนาดสินค้า 12 x 11.5 ซม.', 20, 189, 'recommend', 0, 2, 2, 'BrandJ', 'https://aws.cmzimg.com/upload/10267/product-images/BB384925/86071177.jpg'),
    (11, 'กล่องกิจกรรม 7 ด้าน', 'มีไฟ มีเสียงเพลง ใส่ถ่าน AA จำนวน 3 ก้อน', 2, 3269, 'notrecommend', 51, 3, 3, 'BrandK', 'https://aws.cmzimg.com/upload/10267/product-images/BB849028M-WL/b23d84ac.jpg'),
    (12, 'บอลกลิ้งเทาเวอร์มังกี้', 'ลูกบอลมีเสียงกระดิ๊งผลิตจากพาสติกปลอดสารพิษ', 4, 299, 'notrecommend', 23, 3, 3, 'BrandL', 'https://aws.cmzimg.com/upload/10267/product-images/BB856045W/08c6d5dd.jpg'),
    (13, 'โต๊ะกระดานแม่เหล็ก', 'โต๊ะกระดานแม่เหล็ก4เขียนลบได้ 4 สี พร้อมปากกาและตัวปั้ม 4 ชิ้น', 7, 199, 'notrecommend', 12, 3, 3, 'BrandM', 'https://aws.cmzimg.com/upload/10267/product-images/BB1008752B/0c65f1ff.jpg'),
    (14, 'กล่องกิจกรรม Activity Box', 'ของเล่นที่รวมกิจกรรมหลากหลายไว้ในกล่องเดียว', 8, 1639, 'notrecommend', 46, 4, 3, 'BrandN', 'https://aws.cmzimg.com/upload/10267/product-images/BB1006510W/16846a04.jpg'),
    (15, 'โทรศัพท์กระต่าย', 'โทรศัพท์กระต่ายมีไฟ-มีเสียง', 9, 349, 'notrecommend', 46, 4, 3, 'BrandO', 'https://aws.cmzimg.com/upload/10267/product-images/BB001737/e00e3a0b.jpg'),
    (16, 'ตัวเลขไม้', '0-9 บวก ลบ คูณ หาร', 2, 199, 'notrecommend', 0, 4, 4, 'BrandP', 'https://aws.cmzimg.com/upload/10267/product-images/FW-1551/f4109157.jpg'),
    (17, 'เขาวงกตไดโนเสาร์ฝึกสมาธิ', 'เขาวงกตไดโนเสาร์ฝึกสมาธิ มีไฟ-เสียง', 7, 199, 'notrecommend', 17, 5, 4, 'BrandQ', 'https://aws.cmzimg.com/upload/10267/product-images/BB957158BC/87f90fd8.jpg'),
    (18, 'รถบังคับวิทยุตีลังกา', 'รถบังคับวิทยุตีลังกา 360 องศา', 8, 379, 'notrecommend', 0, 5, 4, 'BrandR', 'https://aws.cmzimg.com/upload/10267/product-images/BB964341W/bd7f721b.jpg'),
    (19, 'เลโก้พิพิธภัณฑ์ไดโนเสาร์', 'เลโก้พิพิธภัณฑ์ไดโนเสาร์ 122 ชิ้น', 19, 159, 'notrecommend', 0, 5, 4, 'BrandS', 'https://aws.cmzimg.com/upload/10267/product-images/BB853766/ea3cfd36.jpg'),
    (20, 'ตัวต่อเลโก้รถแข่ง', 'ตัวต่อเลโก้รถแข่ง มีให้เลือกสะสม 4 สี / 4 แบบ.', 12, 249, 'notrecommend', 15, 5, 4, 'BrandT', 'https://aws.cmzimg.com/upload/10267/product-images/BB333787/0d26a77d.jpg');


CREATE TABLE inventory (
    product_id INT PRIMARY KEY,
    quantity INT,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO inventory (product_id, quantity)
SELECT product_id, product_stock
FROM products;

COMMIT;

-- สร้าง ENUM สำหรับ user_status และ user_role
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_status') THEN
        CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('customer', 'seller', 'admin');
    END IF;
END$$;

-- สร้างตาราง users
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    google_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255), -- เพิ่มคอลัมน์นี้สำหรับชื่อที่ปรับเปลี่ยน
    address TEXT,
    phone VARCHAR(20),
    profile_picture_url VARCHAR(255),
    email_verified BOOLEAN DEFAULT FALSE,
    status user_status NOT NULL DEFAULT 'active',
    role user_role NOT NULL DEFAULT 'customer',
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- สร้างตาราง user_sessions
CREATE TABLE IF NOT EXISTS user_sessions (
    session_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    id_token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- สร้างตาราง user_login_history
CREATE TABLE IF NOT EXISTS user_login_history (
    login_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    login_timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- สร้างตาราง api_keys
CREATE TABLE IF NOT EXISTS api_keys (
    key_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_key VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- สร้าง Trigger สำหรับตาราง users
CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- สร้าง Trigger สำหรับตาราง user_sessions
CREATE TRIGGER update_user_sessions_updated_at
BEFORE UPDATE ON user_sessions
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- สร้าง Trigger สำหรับตาราง api_keys
CREATE TRIGGER update_api_keys_updated_at
BEFORE UPDATE ON api_keys
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- สร้าง Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_google_id ON users(google_id);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_login_history_user_id ON user_login_history(user_id);
CREATE INDEX idx_api_keys_api_key ON api_keys(api_key);
