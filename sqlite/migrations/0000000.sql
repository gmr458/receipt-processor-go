DROP TABLE IF EXISTS "receipt";
DROP TABLE IF EXISTS "item";

CREATE TABLE "receipt" (
	"id"            TEXT NOT NULL,
	"retailer"      TEXT NOT NULL,
	"purchase_date" TEXT NOT NULL,
	"purchase_time" TEXT NOT NULL,
    "total"         REAL NOT NULL,
    
    PRIMARY KEY("id")
);

CREATE TABLE "item" (
	"id"                TEXT NOT NULL,
	"short_description" TEXT NOT NULL,
    "price"             REAL NOT NULL,
	"receipt_id"        TEXT NOT NULL,
    
    FOREIGN KEY("receipt_id") REFERENCES "receipt"("id")
);

