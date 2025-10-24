-- Create "urls" table
CREATE TABLE "urls" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "original_url" text NOT NULL,
  "short_code" character varying(10) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "click_count" bigint NOT NULL DEFAULT 0,
  PRIMARY KEY ("id")
);
-- Create index "idx_urls_short_code" to table: "urls"
CREATE UNIQUE INDEX "idx_urls_short_code" ON "urls" ("short_code");
