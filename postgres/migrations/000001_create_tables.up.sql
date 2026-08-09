CREATE TABLE IF NOT EXISTS "receipt" (
	"id" UUID PRIMARY KEY,
	"retailer" TEXT NOT NULL,
	"purchase_date" DATE NOT NULL,
	"purchase_time" TIME NOT NULL,
	"total" NUMERIC(10,2) NOT NULL,

	CONSTRAINT "receipt_total_non_negative" CHECK ("total" >= 0.0)
);

CREATE TABLE IF NOT EXISTS "item" (
	"id" UUID PRIMARY KEY,
	"short_description" TEXT NOT NULL,
	"price" NUMERIC(10,2) NOT NULL,
	"receipt_id" UUID NOT NULL,

	CONSTRAINT "item_receipt_fk" FOREIGN KEY("receipt_id") REFERENCES "receipt"("id"),

	CONSTRAINT "item_price_non_negative" CHECK ("price" >= 0.0)
);