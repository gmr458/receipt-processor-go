CREATE TABLE "receipt" (
	"id" TEXT NOT NULL,
	"retailer" TEXT NOT NULL,
	"purchase_date" TEXT NOT NULL,
	"purchase_time" TEXT NOT NULL,
	"total" INTEGER NOT NULL,

	CONSTRAINT "receipt_total_non_negative" CHECK ("total" >= 0),

    PRIMARY KEY("id")
);

CREATE TABLE "item" (
	"id" TEXT NOT NULL,
	"short_description" TEXT NOT NULL,
	"price" INTEGER NOT NULL,
	"receipt_id" TEXT NOT NULL,

	CONSTRAINT "item_price_non_negative" CHECK ("price" >= 0),

    FOREIGN KEY("receipt_id") REFERENCES "receipt"("id")
);

CREATE INDEX IF NOT EXISTS "idx_item_receipt_id" ON "item" ("receipt_id");
