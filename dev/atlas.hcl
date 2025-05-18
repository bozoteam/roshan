env "postgres" {
  src = "file://db/schema.hcl"
  url = "postgres://postgres:postgres@localhost:5432/roshan?sslmode=disable"
  dev = "docker://postgres/14"
  migration {
    dir = "file://db/migrations"
  }
}
