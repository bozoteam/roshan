-- Create "user" table
CREATE TABLE "public"."user" (
  "id" uuid NOT NULL,
  "email" character varying(255) NOT NULL,
  "name" character varying(255) NOT NULL,
  "password" character varying(60) NOT NULL,
  "refresh_token" character varying(1024) NULL,
  "created_at" timestamp NOT NULL DEFAULT now (),
  "updated_at" timestamp NOT NULL DEFAULT now (),
  PRIMARY KEY ("id"),
  CONSTRAINT "unique_email" UNIQUE ("email")
);
